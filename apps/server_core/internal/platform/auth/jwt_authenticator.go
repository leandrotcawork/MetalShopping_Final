package auth

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"
)

type JWTAuthenticator struct {
	issuer     string
	audience   string
	algorithm  string
	hmacSecret []byte
	publicKey  *rsa.PublicKey
	now        func() time.Time
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
}

type jwtClaims struct {
	Subject   string  `json:"sub"`
	Issuer    string  `json:"iss"`
	Audience  any     `json:"aud"`
	Expiry    float64 `json:"exp"`
	NotBefore float64 `json:"nbf"`
	TenantID  string  `json:"tenant_id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
}

func NewJWTAuthenticatorFromEnv() (*JWTAuthenticator, error) {
	algorithm := strings.ToUpper(strings.TrimSpace(os.Getenv("MS_AUTH_JWT_ALGORITHM")))
	if algorithm == "" {
		algorithm = "HS256"
	}

	authenticator := &JWTAuthenticator{
		issuer:    strings.TrimSpace(os.Getenv("MS_AUTH_JWT_ISSUER")),
		audience:  strings.TrimSpace(os.Getenv("MS_AUTH_JWT_AUDIENCE")),
		algorithm: algorithm,
		now:       func() time.Time { return time.Now().UTC() },
	}

	switch algorithm {
	case "HS256":
		secret := strings.TrimSpace(os.Getenv("MS_AUTH_JWT_HMAC_SECRET"))
		if secret == "" {
			return nil, fmt.Errorf("auth jwt hmac secret is required")
		}
		authenticator.hmacSecret = []byte(secret)
	case "RS256":
		publicKey, err := parseRSAPublicKey(strings.TrimSpace(os.Getenv("MS_AUTH_JWT_PUBLIC_KEY_PEM")))
		if err != nil {
			return nil, err
		}
		authenticator.publicKey = publicKey
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidAuthenticationMode, algorithm)
	}

	return authenticator, nil
}

func (a *JWTAuthenticator) Authenticate(_ context.Context, accessToken string) (Principal, error) {
	parts := strings.Split(strings.TrimSpace(accessToken), ".")
	if len(parts) != 3 {
		return Principal{}, ErrUnauthenticated
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Principal{}, ErrUnauthenticated
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Principal{}, ErrUnauthenticated
	}

	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return Principal{}, ErrUnauthenticated
	}
	if !strings.EqualFold(header.Algorithm, a.algorithm) {
		return Principal{}, ErrUnauthenticated
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Principal{}, ErrUnauthenticated
	}
	if err := a.verifySignature([]byte(parts[0]+"."+parts[1]), signature); err != nil {
		return Principal{}, ErrUnauthenticated
	}

	var claims jwtClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return Principal{}, ErrUnauthenticated
	}
	if err := a.validateClaims(claims); err != nil {
		return Principal{}, ErrUnauthenticated
	}

	principal := Principal{
		SubjectID: strings.TrimSpace(claims.Subject),
		TenantID:  strings.TrimSpace(claims.TenantID),
		Email:     strings.TrimSpace(claims.Email),
		Name:      strings.TrimSpace(claims.Name),
	}
	if err := principal.Validate(); err != nil {
		return Principal{}, ErrUnauthenticated
	}
	return principal, nil
}

func (a *JWTAuthenticator) verifySignature(message, signature []byte) error {
	hashed := sha256.Sum256(message)
	switch a.algorithm {
	case "HS256":
		mac := hmac.New(sha256.New, a.hmacSecret)
		_, _ = mac.Write(message)
		if !hmac.Equal(mac.Sum(nil), signature) {
			return ErrUnauthenticated
		}
		return nil
	case "RS256":
		return rsa.VerifyPKCS1v15(a.publicKey, crypto.SHA256, hashed[:], signature)
	default:
		return ErrUnauthenticated
	}
}

func (a *JWTAuthenticator) validateClaims(claims jwtClaims) error {
	now := float64(a.now().Unix())
	if strings.TrimSpace(claims.Subject) == "" {
		return ErrUnauthenticated
	}
	if a.issuer != "" && claims.Issuer != a.issuer {
		return ErrUnauthenticated
	}
	if a.audience != "" && !audienceContains(claims.Audience, a.audience) {
		return ErrUnauthenticated
	}
	if claims.Expiry != 0 && now > claims.Expiry {
		return ErrUnauthenticated
	}
	if claims.NotBefore != 0 && now < claims.NotBefore {
		return ErrUnauthenticated
	}
	return nil
}

func audienceContains(value any, expected string) bool {
	switch typed := value.(type) {
	case string:
		return typed == expected
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok && text == expected {
				return true
			}
		}
	}
	return false
}

func parseRSAPublicKey(raw string) (*rsa.PublicKey, error) {
	if raw == "" {
		return nil, fmt.Errorf("auth jwt public key pem is required")
	}
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, fmt.Errorf("auth jwt public key pem is invalid")
	}
	if certificate, err := x509.ParseCertificate(block.Bytes); err == nil {
		if key, ok := certificate.PublicKey.(*rsa.PublicKey); ok {
			return key, nil
		}
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse jwt rsa public key: %w", err)
	}
	key, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("jwt public key is not rsa")
	}
	return key, nil
}
