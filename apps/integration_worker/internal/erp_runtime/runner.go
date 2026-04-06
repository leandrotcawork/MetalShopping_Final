package erp_runtime

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"metalshopping/integration_worker/internal/erp_runtime/raw"
	"metalshopping/integration_worker/internal/erp_runtime/reconciliation"
	"metalshopping/integration_worker/internal/erp_runtime/review"
	"metalshopping/integration_worker/internal/erp_runtime/runs"
	"metalshopping/integration_worker/internal/erp_runtime/staging"
	"metalshopping/integration_worker/internal/erp_runtime/types"
)

// Runner orchestrates the full per-run pipeline:
//  1. Extract raw records from connector
//  2. Save to raw store
//  3. Normalize to staging
//  4. Reconcile staging records
//  5. Create review items for non-promotable records
//  6. Finalize run (update ledger with counts)
//
// NO promotion step — that is server_core's responsibility.
type Runner struct {
	registry    *Registry
	rawStore    *raw.Store
	normalizer  *staging.Normalizer
	reconciler  *reconciliation.Reconciler
	reviewStore *review.Store
	ledger      *runs.Ledger
}

// NewRunner constructs a Runner wiring all pipeline components.
func NewRunner(
	registry *Registry,
	rawStore *raw.Store,
	normalizer *staging.Normalizer,
	reconciler *reconciliation.Reconciler,
	reviewStore *review.Store,
	ledger *runs.Ledger,
) *Runner {
	return &Runner{
		registry:    registry,
		rawStore:    rawStore,
		normalizer:  normalizer,
		reconciler:  reconciler,
		reviewStore: reviewStore,
		ledger:      ledger,
	}
}

// entityCounts accumulates pipeline outcome statistics.
type entityCounts struct {
	promoted int
	warnings int
	rejected int
	reviews  int
}

// Execute runs the full pipeline for a claimed run.
// On per-entity error: continues with remaining entities and marks the run as partial.
func (r *Runner) Execute(ctx context.Context, claim *runs.RunClaim, connectionRef string) error {
	connector, err := r.registry.Get(claim.ConnectorType)
	if err != nil {
		markErr := r.ledger.MarkFailed(ctx, claim.RunID, fmt.Sprintf("connector not found: %v", err))
		if markErr != nil {
			log.Printf("runner: failed to mark run %s as failed: %v", claim.RunID, markErr)
		}
		return err
	}

	var total entityCounts
	var entityErrors []string

	for _, entityName := range claim.EntityScope {
		entity := types.EntityType(entityName)
		counts, err := r.processEntity(ctx, claim, connectionRef, connector, entity)
		if err != nil {
			log.Printf("runner: entity %s error in run %s: %v", entityName, claim.RunID, err)
			entityErrors = append(entityErrors, fmt.Sprintf("%s: %v", entityName, err))
			continue
		}
		total.promoted += counts.promoted
		total.warnings += counts.warnings
		total.rejected += counts.rejected
		total.reviews += counts.reviews
	}

	if len(entityErrors) > 0 {
		summary := strings.Join(entityErrors, "; ")
		return r.ledger.MarkPartial(ctx, claim.RunID, summary,
			total.promoted, total.warnings, total.rejected, total.reviews)
	}
	return r.ledger.MarkCompleted(ctx, claim.RunID,
		total.promoted, total.warnings, total.rejected, total.reviews)
}

// processEntity runs the extract→raw→staging→reconcile→review pipeline for a single entity.
func (r *Runner) processEntity(
	ctx context.Context,
	claim *runs.RunClaim,
	connectionRef string,
	connector Connector,
	entity types.EntityType,
) (entityCounts, error) {
	var counts entityCounts
	var cursor *string

	for {
		req := types.ExtractRequest{
			TenantID:   claim.TenantID,
			RunID:      claim.RunID,
			Entity:     entity,
			Cursor:     cursor,
			Connection: legacyExtractConnection(connectionRef),
		}

		result, err := connector.Extract(ctx, req)
		if err != nil {
			return counts, fmt.Errorf("extract: %w", err)
		}

		if len(result.Records) == 0 {
			break
		}

		// Save raw records
		savedRaw, err := r.rawStore.Save(ctx, claim.TenantID, claim.RunID, result.Records)
		if err != nil {
			return counts, fmt.Errorf("raw save: %w", err)
		}

		// Normalize to staging
		stagingRecords, err := r.normalizer.Normalize(ctx, claim.TenantID, claim.RunID, savedRaw, connector)
		if err != nil {
			return counts, fmt.Errorf("normalize: %w", err)
		}

		// Reconcile
		reconResults, err := r.reconciler.Reconcile(ctx, claim.TenantID, claim.RunID, stagingRecords)
		if err != nil {
			return counts, fmt.Errorf("reconcile: %w", err)
		}

		// Create review items for non-promotable records
		if err := r.reviewStore.CreateFromReconciliation(ctx, claim.InstanceID, claim.ConnectorType, reconResults, stagingRecords); err != nil {
			return counts, fmt.Errorf("review items: %w", err)
		}

		// Tally counts
		for _, rr := range reconResults {
			switch rr.Classification {
			case reconciliation.ClassificationPromotable:
				counts.promoted++
			case reconciliation.ClassificationPromotableWithWarning:
				counts.promoted++
				counts.warnings++
			case reconciliation.ClassificationReviewRequired:
				counts.reviews++
			case reconciliation.ClassificationRejected:
				counts.rejected++
				counts.reviews++
			}
		}

		if !result.HasMore {
			break
		}
		cursor = result.NextCursor
	}

	return counts, nil
}

func legacyExtractConnection(connectionRef string) types.ExtractConnection {
	ref := strings.TrimSpace(connectionRef)
	if ref == "" {
		return types.ExtractConnection{Kind: "oracle"}
	}

	u, err := url.Parse(ref)
	if err != nil {
		return types.ExtractConnection{
			Kind: "oracle",
			Host: ref,
		}
	}

	if u.Scheme == "fixture" {
		return types.ExtractConnection{
			Kind: "oracle",
			Host: "fixture",
		}
	}

	connection := types.ExtractConnection{
		Kind: "oracle",
		Host: strings.TrimSpace(u.Hostname()),
		Port: 1521,
	}
	if rawPort := strings.TrimSpace(u.Port()); rawPort != "" {
		if parsedPort, convErr := strconv.Atoi(rawPort); convErr == nil && parsedPort > 0 {
			connection.Port = parsedPort
		}
	}
	if u.User != nil {
		connection.Username = strings.TrimSpace(u.User.Username())
		if password, ok := u.User.Password(); ok {
			connection.PasswordSecretRef = strings.TrimSpace(password)
		}
	}
	if connection.Username == "" {
		connection.Username = strings.TrimSpace(u.Query().Get("user"))
	}
	if connection.Username == "" {
		connection.Username = strings.TrimSpace(u.Query().Get("username"))
	}
	if connection.PasswordSecretRef == "" {
		connection.PasswordSecretRef = strings.TrimSpace(u.Query().Get("password"))
	}
	if service := strings.TrimSpace(u.Query().Get("service")); service != "" {
		connection.ServiceName = &service
	} else if serviceName := strings.TrimSpace(u.Query().Get("serviceName")); serviceName != "" {
		connection.ServiceName = &serviceName
	} else if sid := strings.TrimSpace(u.Query().Get("sid")); sid != "" {
		connection.SID = &sid
	}

	return connection
}
