package authsession

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrCSRFMalformed = errors.New("csrf token is malformed")
	ErrCSRFBadMAC    = errors.New("csrf token signature is invalid")
	ErrCSRFTimedOut  = errors.New("csrf token expired")
)

type csrfManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func newCSRFManager(config Config) *csrfManager {
	return &csrfManager{
		secret: []byte(config.CSRFHMACSecret),
		ttl:    config.CSRFTTL,
		now:    defaultNow,
	}
}

func (m *csrfManager) Issue(sessionID string, tenantID string) (string, time.Time, error) {
	if m == nil || len(m.secret) == 0 {
		return "", time.Time{}, ErrOIDCConfigIncomplete
	}

	nonce, err := randomToken(18)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate csrf nonce: %w", err)
	}

	issuedAt := truncateTime(m.now())
	expiresAt := issuedAt.Add(m.ttl)
	payload := csrfPayload(sessionID, tenantID, nonce, issuedAt.Unix())
	signature := signCSRF(m.secret, payload)

	token := strings.Join([]string{
		nonce,
		strconv.FormatInt(issuedAt.Unix(), 10),
		signature,
	}, ".")
	return token, expiresAt, nil
}

func (m *csrfManager) Validate(token string, sessionID string, tenantID string) error {
	if m == nil || len(m.secret) == 0 {
		return ErrOIDCConfigIncomplete
	}

	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return ErrCSRFMalformed
	}

	nonce := strings.TrimSpace(parts[0])
	rawIssuedAt := strings.TrimSpace(parts[1])
	signature := strings.TrimSpace(parts[2])
	if nonce == "" || rawIssuedAt == "" || signature == "" {
		return ErrCSRFMalformed
	}

	issuedAtUnix, err := strconv.ParseInt(rawIssuedAt, 10, 64)
	if err != nil {
		return ErrCSRFMalformed
	}

	payload := csrfPayload(sessionID, tenantID, nonce, issuedAtUnix)
	expectedSignature := signCSRF(m.secret, payload)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return ErrCSRFBadMAC
	}

	issuedAt := time.Unix(issuedAtUnix, 0).UTC()
	if m.now().UTC().After(issuedAt.Add(m.ttl)) {
		return ErrCSRFTimedOut
	}

	return nil
}

func csrfPayload(sessionID string, tenantID string, nonce string, issuedAtUnix int64) string {
	return strings.Join([]string{
		strings.TrimSpace(sessionID),
		strings.TrimSpace(tenantID),
		strings.TrimSpace(nonce),
		strconv.FormatInt(issuedAtUnix, 10),
	}, "|")
}

func signCSRF(secret []byte, payload string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
