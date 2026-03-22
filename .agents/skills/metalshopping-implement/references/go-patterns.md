# Go Patterns

## Nullable columns
```go
var finishedAt sql.NullTime
rows.Scan(..., &finishedAt)
if finishedAt.Valid {
    value := finishedAt.Time.UTC()
    item.FinishedAt = &value
}
```

## List with pagination
```go
const countQ = `SELECT COUNT(*) FROM t WHERE tenant_id = current_tenant_id() AND ($1='all' OR status=$1)`
var total int64
tx.QueryRowContext(ctx, countQ, filter).Scan(&total)

const listQ = `SELECT ... FROM t WHERE tenant_id = current_tenant_id() ORDER BY created_at DESC LIMIT $1 OFFSET $2`
rows, _ := tx.QueryContext(ctx, listQ, limit, offset)
```

## UUID v4 generation
```go
func generateID() string {
    buf := make([]byte, 16)
    rand.Read(buf)
    buf[6] = (buf[6] & 0x0f) | 0x40
    buf[8] = (buf[8] & 0x3f) | 0x80
    h := hex.EncodeToString(buf)
    return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}
```

## Error wrapping
```go
// always wrap with context on every layer crossing
return fmt.Errorf("query analytics overview for tenant %s: %w", tenantID, err)

// domain sentinel errors
var ErrRunNotFound = errors.New("shopping run not found")

// handler checks
if errors.Is(err, postgres.ErrRunNotFound) {
    writeError(w, 404, "SHOPPING_RUN_NOT_FOUND", "Shopping run not found", traceID)
    return
}
```

## Governance guard
```go
// ports/repository.go — define the interface
type XCreationGuard interface {
    IsXCreationEnabled(ctx context.Context, tenantID string) (bool, error)
}

// adapters/governance/x_guard.go — implement
type XCreationGuard struct { resolver *feature_flags.Resolver; environment string }
func (g *XCreationGuard) IsXCreationEnabled(_ context.Context, tenantID string) (bool, error) {
    return g.resolver.Resolve(bootstrap.XCreationEnabledKey, feature_flags.ResolutionContext{
        Environment: g.environment, TenantID: tenantID,
    })
}
// Add const XCreationEnabledKey = "x.creation.enabled" to bootstrap/bootstrap.go
```
