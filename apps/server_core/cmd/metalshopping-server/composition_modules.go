package main

import (
	"context"
	"log"
	"net/http"

	cataloggov "metalshopping/server_core/internal/modules/catalog/adapters/governance"
	catalogpg "metalshopping/server_core/internal/modules/catalog/adapters/postgres"
	catalogapp "metalshopping/server_core/internal/modules/catalog/application"
	catalogreadmodel "metalshopping/server_core/internal/modules/catalog/readmodel"
	cataloghttp "metalshopping/server_core/internal/modules/catalog/transport/http"
	homepg "metalshopping/server_core/internal/modules/home/adapters/postgres"
	homeapp "metalshopping/server_core/internal/modules/home/application"
	homehttp "metalshopping/server_core/internal/modules/home/transport/http"
	iamgov "metalshopping/server_core/internal/modules/iam/adapters/governance"
	iampg "metalshopping/server_core/internal/modules/iam/adapters/postgres"
	iamapp "metalshopping/server_core/internal/modules/iam/application"
	iamhttp "metalshopping/server_core/internal/modules/iam/transport/http"
	inventorypg "metalshopping/server_core/internal/modules/inventory/adapters/postgres"
	inventoryapp "metalshopping/server_core/internal/modules/inventory/application"
	inventoryhttp "metalshopping/server_core/internal/modules/inventory/transport/http"
	pricinggov "metalshopping/server_core/internal/modules/pricing/adapters/governance"
	pricingpg "metalshopping/server_core/internal/modules/pricing/adapters/postgres"
	pricingapp "metalshopping/server_core/internal/modules/pricing/application"
	pricinghttp "metalshopping/server_core/internal/modules/pricing/transport/http"
	shoppingpg "metalshopping/server_core/internal/modules/shopping/adapters/postgres"
	shoppingapp "metalshopping/server_core/internal/modules/shopping/application"
	shoppinghttp "metalshopping/server_core/internal/modules/shopping/transport/http"
	supplierspg "metalshopping/server_core/internal/modules/suppliers/adapters/postgres"
	suppliersapp "metalshopping/server_core/internal/modules/suppliers/application"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

type moduleComposition struct {
	iamRepo       *iampg.Repository
	iamAuthorizer *iamapp.StaticAuthorizer
	registerHTTP  func(mux *http.ServeMux)
}

func composeModules(ctx context.Context, runtime runtimeComposition, governance governanceComposition) moduleComposition {
	if ctx == nil {
		ctx = context.Background()
	}

	outboxStore := outbox.NewStore(runtime.db)
	outboxDispatcher := outbox.NewDispatcher(outboxStore, outbox.NewLoggingPublisher(log.Default()))
	go outboxDispatcher.Run(ctx)

	iamRepo := iampg.NewRepository(runtime.db)
	catalogRepo := catalogpg.NewRepository(runtime.db, outboxStore)
	inventoryRepo := inventorypg.NewRepository(runtime.db, outboxStore)
	pricingRepo := pricingpg.NewRepository(runtime.db, outboxStore)
	homeSummaryReader := homepg.NewSummaryReader(runtime.db)
	suppliersDirectoryReader := supplierspg.NewDirectoryReader(runtime.db)
	suppliersService := suppliersapp.NewService(suppliersDirectoryReader)
	shoppingReader := shoppingpg.NewReader(runtime.db, suppliersService)
	shoppingWriter := shoppingpg.NewWriter(runtime.db)

	iamAuthorizer := iamapp.NewStaticAuthorizer()
	iamAuthorization := iamapp.NewAuthorizationService(iamRepo, iamAuthorizer)

	iamAdminService := iamapp.NewAdminService(
		iamRepo,
		iamgov.NewAdminPolicyGuard(governance.policies, runtime.environment),
	)
	iamAdminHandler := iamhttp.NewAdminHandler(iamAdminService, iamAuthorization)

	catalogProductCreationGuard := cataloggov.NewProductCreationGuard(governance.featureFlags, runtime.environment)
	catalogDescriptionGuard := cataloggov.NewDescriptionGuard(governance.thresholds, runtime.environment)
	catalogService := catalogapp.NewService(catalogRepo, catalogProductCreationGuard, catalogDescriptionGuard)
	catalogProductsPortfolioService := catalogreadmodel.NewProductsPortfolioService(catalogRepo)
	catalogHandler := cataloghttp.NewHandler(catalogService, catalogProductsPortfolioService, iamAuthorization)

	inventoryService := inventoryapp.NewService(inventoryRepo)
	inventoryHandler := inventoryhttp.NewHandler(inventoryService, iamAuthorization)

	pricingManualOverrideGuard := pricinggov.NewManualOverrideGuard(governance.policies, runtime.environment)
	pricingService := pricingapp.NewService(pricingRepo, pricingManualOverrideGuard)
	pricingHandler := pricinghttp.NewHandler(pricingService, iamAuthorization)
	homeService := homeapp.NewService(homeSummaryReader)
	homeHandler := homehttp.NewHandler(homeService)
	shoppingService := shoppingapp.NewService(shoppingReader, shoppingWriter)
	shoppingHandler := shoppinghttp.NewHandler(shoppingService)

	return moduleComposition{
		iamRepo:       iamRepo,
		iamAuthorizer: iamAuthorizer,
		registerHTTP: func(mux *http.ServeMux) {
			iamAdminHandler.RegisterRoutes(mux)
			catalogHandler.RegisterRoutes(mux)
			inventoryHandler.RegisterRoutes(mux)
			pricingHandler.RegisterRoutes(mux)
			homeHandler.RegisterRoutes(mux)
			shoppingHandler.RegisterRoutes(mux)
		},
	}
}
