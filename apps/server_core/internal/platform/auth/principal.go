package auth

import (
	"fmt"
	"strings"
)

type Principal struct {
	SubjectID string
	TenantID  string
	Email     string
	Name      string
}

func (p Principal) Validate() error {
	if strings.TrimSpace(p.SubjectID) == "" {
		return fmt.Errorf("auth principal missing subject id")
	}
	return nil
}
