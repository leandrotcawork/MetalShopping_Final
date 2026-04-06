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
	entityStepStore := runs.NewEntityStepStore(db)
	runner := erp_runtime.NewRunner(registry, rawStore, normalizer, reconciler, reviewStore, ledger, entityStepStore)

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

		// Fetch structured connection config for this instance
		connection, err := getConnection(tenantCtx, db, claim.TenantID, claim.InstanceID)
		if err != nil {
			log.Printf("erp-sync: getConnection error for instance %s: %v", claim.InstanceID, err)
			if markErr := ledger.MarkFailed(tenantCtx, claim.RunID, fmt.Sprintf("connection lookup failed: %v", err)); markErr != nil {
				log.Printf("erp-sync: MarkFailed error: %v", markErr)
			}
			continue
		}

		// Execute in a goroutine so the claim loop can pick up the next run immediately
		go func(c *runs.RunClaim, conn erp_runtime.ExtractConnection, runCtx context.Context) {
			if execErr := runner.Execute(runCtx, c, conn); execErr != nil {
				log.Printf("erp-sync: run %s failed: %v", c.RunID, execErr)
			} else {
				log.Printf("erp-sync: run %s completed", c.RunID)
			}
		}(claim, connection, tenantCtx)
	}
}

// getConnection queries erp_integration_instances for structured Oracle connection fields.
func getConnection(ctx context.Context, db *sql.DB, tenantID, instanceID string) (erp_runtime.ExtractConnection, error) {
	tx, err := tenantdb.BeginTenantTx(ctx, db, tenantID, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return erp_runtime.ExtractConnection{}, err
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
SELECT connection_kind, db_host, db_port, db_service_name, db_sid, db_username,
       db_password_secret_ref, connect_timeout_seconds, fetch_batch_size, entity_batch_size
FROM erp_integration_instances
WHERE instance_id = $1
  AND tenant_id = current_tenant_id()`
	var conn erp_runtime.ExtractConnection
	var serviceName sql.NullString
	var sid sql.NullString
	if err := tx.QueryRowContext(ctx, q, instanceID).Scan(
		&conn.Kind,
		&conn.Host,
		&conn.Port,
		&serviceName,
		&sid,
		&conn.Username,
		&conn.PasswordSecretRef,
		&conn.ConnectTimeoutSec,
		&conn.FetchBatchSize,
		&conn.EntityBatchSize,
	); err != nil {
		return erp_runtime.ExtractConnection{}, fmt.Errorf("instance %s: %w", instanceID, err)
	}
	if serviceName.Valid {
		conn.ServiceName = &serviceName.String
	}
	if sid.Valid {
		conn.SID = &sid.String
	}
	if err := tx.Commit(); err != nil {
		return erp_runtime.ExtractConnection{}, err
	}
	return conn, nil
}

// sleep waits for the given duration or until ctx is cancelled.
func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
