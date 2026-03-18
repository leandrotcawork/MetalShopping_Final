package authsession

import (
	"context"
	"fmt"
	"time"

	governancebootstrap "metalshopping/server_core/internal/platform/governance/bootstrap"
	"metalshopping/server_core/internal/platform/governance/feature_flags"
	"metalshopping/server_core/internal/platform/governance/threshold_resolver"
)

type RuntimePolicy struct {
	Enabled         bool
	IdleTimeout     time.Duration
	AbsoluteTimeout time.Duration
}

type RuntimePolicyResolver struct {
	featureFlags *feature_flags.Resolver
	thresholds   *threshold_resolver.Resolver
	environment  string
}

func NewRuntimePolicyResolver(featureFlags *feature_flags.Resolver, thresholds *threshold_resolver.Resolver, environment string) *RuntimePolicyResolver {
	return &RuntimePolicyResolver{
		featureFlags: featureFlags,
		thresholds:   thresholds,
		environment:  environment,
	}
}

func (r *RuntimePolicyResolver) Resolve(_ context.Context, tenantID string) (RuntimePolicy, error) {
	if r == nil || r.featureFlags == nil || r.thresholds == nil {
		return RuntimePolicy{}, ErrGovernanceResolution
	}

	enabled, err := r.featureFlags.Resolve(governancebootstrap.AuthWebSessionEnabledKey, feature_flags.ResolutionContext{
		Environment: r.environment,
		TenantID:    tenantID,
		Module:      "auth",
	})
	if err != nil {
		return RuntimePolicy{}, fmt.Errorf("%w: resolve web session flag: %v", ErrGovernanceResolution, err)
	}
	if !enabled {
		return RuntimePolicy{Enabled: false}, nil
	}

	idleMinutes, found, err := r.thresholds.Resolve(governancebootstrap.AuthSessionIdleTimeoutMinutesKey, threshold_resolver.ResolutionContext{
		Environment: r.environment,
		TenantID:    tenantID,
		Module:      "auth",
	})
	if err != nil {
		return RuntimePolicy{}, fmt.Errorf("%w: resolve idle timeout: %v", ErrGovernanceResolution, err)
	}
	if !found {
		return RuntimePolicy{}, ErrMissingSessionTimeout
	}

	absoluteMinutes, found, err := r.thresholds.Resolve(governancebootstrap.AuthSessionAbsoluteTimeoutMinutesKey, threshold_resolver.ResolutionContext{
		Environment: r.environment,
		TenantID:    tenantID,
		Module:      "auth",
	})
	if err != nil {
		return RuntimePolicy{}, fmt.Errorf("%w: resolve absolute timeout: %v", ErrGovernanceResolution, err)
	}
	if !found {
		return RuntimePolicy{}, ErrMissingSessionTimeout
	}

	return RuntimePolicy{
		Enabled:         true,
		IdleTimeout:     time.Duration(idleMinutes * float64(time.Minute)),
		AbsoluteTimeout: time.Duration(absoluteMinutes * float64(time.Minute)),
	}, nil
}
