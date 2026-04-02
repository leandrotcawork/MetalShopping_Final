package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	erp_runtime "metalshopping/integration_worker/internal/erp_runtime"
	"metalshopping/integration_worker/internal/erp_runtime/connectors/sankhya"
	"metalshopping/integration_worker/internal/erp_runtime/raw"
	"metalshopping/integration_worker/internal/erp_runtime/reconciliation"
	"metalshopping/integration_worker/internal/erp_runtime/review"
	"metalshopping/integration_worker/internal/erp_runtime/runs"
	"metalshopping/integration_worker/internal/erp_runtime/staging"
	"metalshopping/integration_worker/internal/erp_runtime/tenantdb"
)

func main() {
	// 1. Connect to Postgres
	dsn := os.Getenv("ERP_SYNC_DATABASE_URL")
	if dsn == "" {
		log.Fatal("erp-sync: ERP_SYNC_DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("erp-sync: open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("erp-sync: ping db: %v", err)
	}
	log.Println("erp-sync: database connected")

	// 2. Build runtime components
	registry := erp_runtime.NewRegistry()
	rawStore := raw.NewStore(db)
	normalizer := staging.NewNormalizer(db)
	reconciler := reconciliation.NewReconciler(db)
	reviewStore := review.NewStore(db)
	ledger := runs.NewLedger(db)
	runner := erp_runtime.NewRunner(registry, rawStore, normalizer, reconciler, reviewStore, ledger)

	// 3. Register connectors
	registry.Register(sankhya.New())
	_ = runner // runner is used in the claim loop below

	// 4. Signal-aware context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("erp-sync: starting run-claim loop")

	// 5. Run-claim loop
	for {
		select {
		case <-ctx.Done():
			log.Println("erp-sync: shutting down")
			return
		default:
		}

		claim, err := ledger.ClaimPendingRun(ctx)
		if err != nil {
			log.Printf("erp-sync: claim error: %v", err)
			sleep(ctx, 5*time.Second)
			continue
		}
		if claim == nil {
			// No pending run available
			sleep(ctx, 5*time.Second)
			continue
		}

		log.Printf("erp-sync: claimed run %s (tenant=%s instance=%s connector=%s)",
			claim.RunID, claim.TenantID, claim.InstanceID, claim.ConnectorType)

		tenantCtx, err := tenantdb.WithTenantID(ctx, claim.TenantID)
		if err != nil {
			log.Printf("erp-sync: invalid tenant context for run %s: %v", claim.RunID, err)
			continue
		}

		// Fetch connection_ref for this instance
		connectionRef, err := getConnectionRef(tenantCtx, db, claim.TenantID, claim.InstanceID)
		if err != nil {
			log.Printf("erp-sync: getConnectionRef error for instance %s: %v", claim.InstanceID, err)
			if markErr := ledger.MarkFailed(tenantCtx, claim.RunID, fmt.Sprintf("connection_ref lookup failed: %v", err)); markErr != nil {
				log.Printf("erp-sync: MarkFailed error: %v", markErr)
			}
			continue
		}

		// Execute in a goroutine so the claim loop can pick up the next run immediately
		go func(c *runs.RunClaim, ref string) {
			if execErr := runner.Execute(tenantCtx, c, ref); execErr != nil {
				log.Printf("erp-sync: run %s failed: %v", c.RunID, execErr)
			} else {
				log.Printf("erp-sync: run %s completed", c.RunID)
			}
		}(claim, connectionRef)
	}
}

// getConnectionRef queries erp_integration_instances for the connection_ref of a given instance.
func getConnectionRef(ctx context.Context, db *sql.DB, tenantID, instanceID string) (string, error) {
	tx, err := tenantdb.BeginTenantTx(ctx, db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
SELECT connection_ref
FROM erp_integration_instances
WHERE instance_id = $1
  AND tenant_id = current_tenant_id()`
	var ref string
	if err := tx.QueryRowContext(ctx, q, instanceID).Scan(&ref); err != nil {
		return "", fmt.Errorf("instance %s: %w", instanceID, err)
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return ref, nil
}

// sleep waits for the given duration or until ctx is cancelled.
func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
