package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

type jwksDocument struct {
	Keys []jwksKey `json:"keys"`
}

type jwksKey struct {
	KeyID string `json:"kid"`
	Kty   string `json:"kty"`
	N     string `json:"n"`
	E     string `json:"e"`
}

func fetchJWKS(client *http.Client, jwksURL string) (map[string]*rsa.PublicKey, error) {
	response, err := client.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch jwks: unexpected status %d", response.StatusCode)
	}

	var document jwksDocument
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(document.Keys))
	for _, key := range document.Keys {
		if key.Kty != "RSA" || key.KeyID == "" {
			continue
		}
		publicKey, err := parseJWKSRSAKey(key.N, key.E)
		if err != nil {
			continue
		}
		keys[key.KeyID] = publicKey
	}
	return keys, nil
}

func parseJWKSRSAKey(modulusBase64 string, exponentBase64 string) (*rsa.PublicKey, error) {
	modulusBytes, err := base64.RawURLEncoding.DecodeString(modulusBase64)
	if err != nil {
		return nil, err
	}
	exponentBytes, err := base64.RawURLEncoding.DecodeString(exponentBase64)
	if err != nil {
		return nil, err
	}

	modulus := new(big.Int).SetBytes(modulusBytes)
	exponent := new(big.Int).SetBytes(exponentBytes)
	return &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}, nil
}
