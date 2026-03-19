package main

import (
	"context"
	"fmt"

	governancebootstrap "metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
	"metalshopping/server_core/internal/platform/governance/policy_resolver"
	"metalshopping/server_core/internal/platform/governance/threshold_resolver"
)

type governanceComposition struct {
	featureFlags *feature_flags.Resolver
	thresholds   *threshold_resolver.Resolver
	policies     *policy_resolver.Resolver
}

func composeGovernance(ctx context.Context, runtime runtimeComposition) (governanceComposition, error) {
	registry := governancebootstrap.NewRegistry()

	featureFlagResolver, err := feature_flags.NewPostgresResolver(ctx, runtime.db, registry)
	if err != nil {
		return governanceComposition{}, fmt.Errorf("load governance feature flags from database: %w", err)
	}

	thresholdResolver, err := threshold_resolver.NewPostgresResolver(ctx, runtime.db, registry)
	if err != nil {
		return governanceComposition{}, fmt.Errorf("load governance thresholds from database: %w", err)
	}

	policyResolver, err := policy_resolver.NewPostgresResolver(ctx, runtime.db, registry)
	if err != nil {
		return governanceComposition{}, fmt.Errorf("load governance policies from database: %w", err)
	}

	return governanceComposition{
		featureFlags: featureFlagResolver,
		thresholds:   thresholdResolver,
		policies:     policyResolver,
	}, nil
}
