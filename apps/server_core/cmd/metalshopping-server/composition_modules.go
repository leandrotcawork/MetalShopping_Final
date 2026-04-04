package main

import (
	"context"
	"log"
	"net/http"

	analyticspg "metalshopping/server_core/internal/modules/analytics_serving/adapters/postgres"
	analyticsapp "metalshopping/server_core/internal/modules/analytics_serving/application"
	analyticshttp "metalshopping/server_core/internal/modules/analytics_serving/transport/http"
	cataloggov "metalshopping/server_core/internal/modules/catalog/adapters/governance"
	catalogpg "metalshopping/server_core/internal/modules/catalog/adapters/postgres"
	catalogapp "metalshopping/server_core/internal/modules/catalog/application"
	catalogreadmodel "metalshopping/server_core/internal/modules/catalog/readmodel"
	cataloghttp "metalshopping/server_core/internal/modules/catalog/transport/http"
	erpcatalog "metalshopping/server_core/internal/modules/erp_integrations/adapters/catalog"
	erpgov "metalshopping/server_core/internal/modules/erp_integrations/adapters/governance"
	erpiam "metalshopping/server_core/internal/modules/erp_integrations/adapters/iam"
	erpinventory "metalshopping/server_core/internal/modules/erp_integrations/adapters/inventory"
	erppg "metalshopping/server_core/internal/modules/erp_integrations/adapters/postgres"
	erppricing "metalshopping/server_core/internal/modules/erp_integrations/adapters/pricing"
	erpapp "metalshopping/server_core/internal/modules/erp_integrations/application"
	erphttp "metalshopping/server_core/internal/modules/erp_integrations/transport/http"
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
	suppliershttp "metalshopping/server_core/internal/modules/suppliers/transport/http"
	"metalshopping/server_core/internal/platform/messaging/outbox"
	platformsuppliers "metalshopping/server_core/internal/platform/suppliers"
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
	analyticsHomeReader := analyticspg.NewReader(runtime.db)
	suppliersRepo := supplierspg.NewRepository(runtime.db)
	suppliersService := suppliersapp.NewService(suppliersRepo, platformsuppliers.NewDefaultRegistry())
	shoppingReader := shoppingpg.NewReader(runtime.db, suppliersService)
	shoppingWriter := shoppingpg.NewWriter(runtime.db, outboxStore)

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
	analyticsService := analyticsapp.NewService(analyticsHomeReader)
	analyticsHandler := analyticshttp.NewHandler(analyticsService)
	shoppingService := shoppingapp.NewService(shoppingReader, shoppingWriter)
	shoppingHandler := shoppinghttp.NewHandler(shoppingService)
	suppliersHandler := suppliershttp.NewHandler(suppliersService)

	// ERP integrations module
	erpRepos := erppg.NewRepos(runtime.db, outboxStore)
	erpPriceWriter := erppricing.NewWriter(pricingService, erpRepos.Reconciliations)
	erpInventoryWriter := erpinventory.NewWriter(inventoryService)
	erpEnabledGuard := erpgov.NewIntegrationEnabledGuard(governance.featureFlags, runtime.environment)
	erpAutoPromoGuard := erpgov.NewAutoPromotionGuard(governance.policies, runtime.environment)
	erpPermChecker := erpiam.NewPermissionChecker(iamAuthorization)
	erpProductWriter := erpcatalog.NewProductWriter(runtime.db, outboxStore, catalogRepo)
	erpProductPromotion := erpapp.NewProductPromotion(erpRepos.Staging, erpRepos.Runs, erpProductWriter)
	erpPricePromotion := erpapp.NewPricePromotion(erpRepos.Staging, erpRepos.Runs, catalogRepo, erpPriceWriter)
	erpInventoryPromotion := erpapp.NewInventoryPromotion(erpRepos.Staging, erpRepos.Runs, catalogRepo, erpInventoryWriter)
	erpSvc := erpapp.NewService(
		erpRepos.Instances,
		erpRepos.Runs,
		erpRepos.Reviews,
		erpEnabledGuard,
		erpPermChecker,
		outboxStore,
	)
	erpHandler := erphttp.NewHandler(erpSvc)
	erpPromoConsumer := erpapp.NewPromotionConsumer(erpRepos.Reconciliations, erpAutoPromoGuard, erpProductPromotion, erpPricePromotion, erpInventoryPromotion)
	go erpPromoConsumer.Start(ctx)

	return moduleComposition{
		iamRepo:       iamRepo,
		iamAuthorizer: iamAuthorizer,
		registerHTTP: func(mux *http.ServeMux) {
			iamAdminHandler.RegisterRoutes(mux)
			catalogHandler.RegisterRoutes(mux)
			inventoryHandler.RegisterRoutes(mux)
			pricingHandler.RegisterRoutes(mux)
			homeHandler.RegisterRoutes(mux)
			analyticsHandler.RegisterRoutes(mux)
			shoppingHandler.RegisterRoutes(mux)
			suppliersHandler.RegisterRoutes(mux)
			erpHandler.RegisterRoutes(mux)
		},
	}
}
