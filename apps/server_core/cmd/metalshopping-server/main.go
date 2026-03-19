package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"metalshopping/server_core/internal/platform/runtime_config"
)

func main() {
	processCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	if err := runtime_config.LoadDotEnvIfPresent(".env"); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	runtime, err := composeRuntime(processCtx)
	if err != nil {
		log.Fatalf("compose runtime: %v", err)
	}
	defer func() { _ = runtime.db.Close() }()

	governance, err := composeGovernance(processCtx, runtime)
	if err != nil {
		log.Fatalf("compose governance: %v", err)
	}

	modules := composeModules(processCtx, runtime, governance)
	authSession, err := composeAuthSession(runtime, governance, modules)
	if err != nil {
		log.Fatalf("compose auth/session: %v", err)
	}
	server := composeHTTPServer(runtime, modules, authSession)

	go func() {
		<-processCtx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http shutdown error: %v", err)
		}
	}()

	log.Printf("metalshopping-server listening on %s", runtime.addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}
