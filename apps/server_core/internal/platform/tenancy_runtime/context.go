package tenancy_runtime

import "context"

type Tenant struct {
	ID string
}

type tenantContextKey struct{}

func WithTenant(ctx context.Context, tenant Tenant) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenant)
}

func TenantFromContext(ctx context.Context) (Tenant, bool) {
	tenant, ok := ctx.Value(tenantContextKey{}).(Tenant)
	return tenant, ok
}
