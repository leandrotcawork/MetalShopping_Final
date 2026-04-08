package oracle

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Config captures the validated Oracle connection parameters.
type Config struct {
	Host              string
	Port              int
	ServiceName       *string
	SID               *string
	Username          string
	Password          string
	ConnectTimeoutSec int
}

// ConnectString validates the config and builds a DSN representation.
func (c Config) ConnectString() (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}

	u := &url.URL{
		Scheme: "oracle",
		User:   url.UserPassword(c.Username, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
	}

	query := u.Query()
	if c.ServiceName != nil {
		u.Path = "/" + strings.TrimSpace(*c.ServiceName)
	} else {
		query.Set("sid", strings.TrimSpace(*c.SID))
	}
	if c.ConnectTimeoutSec > 0 {
		query.Set("connect_timeout", strconv.Itoa(c.ConnectTimeoutSec))
	}
	if encoded := query.Encode(); encoded != "" {
		u.RawQuery = encoded
	}
	return u.String(), nil
}

func (c Config) validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("oracle config host must not be empty")
	}
	if c.Port <= 0 {
		return fmt.Errorf("oracle config port must be a positive integer")
	}
	hasServiceName := c.ServiceName != nil && strings.TrimSpace(*c.ServiceName) != ""
	hasSID := c.SID != nil && strings.TrimSpace(*c.SID) != ""
	switch {
	case hasServiceName && hasSID:
		return fmt.Errorf("oracle config must set exactly one of service_name or sid")
	case !hasServiceName && !hasSID:
		return fmt.Errorf("oracle config must set exactly one of service_name or sid")
	}
	if strings.TrimSpace(c.Username) == "" {
		return fmt.Errorf("oracle config username must not be empty")
	}
	if strings.TrimSpace(c.Password) == "" {
		return fmt.Errorf("oracle config password must not be empty")
	}
	if c.ConnectTimeoutSec < 0 {
		return fmt.Errorf("oracle config connect_timeout_sec must not be negative")
	}
	return nil
}
