package unit

import (
	"testing"

	"metalshopping/server_core/internal/modules/iam/application"
	"metalshopping/server_core/internal/modules/iam/domain"
)

func TestStaticAuthorizerHonorsExpectedPermissions(t *testing.T) {
	authorizer := application.NewStaticAuthorizer()

	if !authorizer.Can(domain.RoleAdmin, domain.PermIAMManageRoles) {
		t.Fatal("expected admin to manage roles")
	}
	if authorizer.Can(domain.RoleViewer, domain.PermPricingWrite) {
		t.Fatal("expected viewer not to write pricing")
	}
	if !authorizer.Can(domain.RoleInventoryManager, domain.PermInventoryWrite) {
		t.Fatal("expected inventory manager to write inventory")
	}
	if !authorizer.Can(domain.RoleAnalyst, domain.PermAnalyticsServingRead) {
		t.Fatal("expected analyst to read analytics serving")
	}
}
