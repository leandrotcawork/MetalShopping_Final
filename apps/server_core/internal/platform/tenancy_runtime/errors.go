package tenancy_runtime

import "errors"

var (
	ErrMissingTenantContext = errors.New("tenancy missing tenant context")
)
