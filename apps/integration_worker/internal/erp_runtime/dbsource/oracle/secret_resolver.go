package oracle

import "context"

// SecretResolver resolves a secret reference to its plaintext value.
type SecretResolver interface {
	Resolve(ctx context.Context, ref string) (string, error)
}
