package main

import (
	"context"
	"fmt"
	"github.com/base-cloud-engine/example-agent/api"
	"github.com/base-cloud-engine/example-agent/internal/app"
	"github.com/base-cloud-engine/example-agent/internal/controllers"
	"github.com/base-cloud-engine/example-agent/internal/repositories/exampleresourcerepo"
	"github.com/base-cloud-engine/example-agent/internal/services/exampleresource"
	"github.com/base-cloud-engine/example-agent/pkg/logger"
	"github.com/reconcile-kit/api/resource"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var Version string

func main() {
	_ = context.Background()
	if syscall.Geteuid() != 0 {
		log.Fatal("error: must be run as root")
	}

	cfg, err := app.ParseConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("config error: %w", err))
	}
	l := logger.New(logger.Level(cfg.LogLevel))
	l.Infof("example-agent version: %s", Version)

	repo := exampleresourcerepo.New()
	exampleResourceService := exampleresource.NewService(repo)

	currentApp := app.New(
		cfg.ShardID,
		cfg.InformerURL,
		cfg.StorageURL,
		l,
	)

	err = app.SetController[*api.ExampleResource](
		currentApp,
		resource.GroupKind{Kind: "example-resource", Group: "central.base-cloud-engine.com"},
		controllers.NewExampleResourceReconciler[*api.ExampleResource](l, exampleResourceService),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctxExit, cancel := context.WithCancel(context.Background())
	err = currentApp.Run(ctxExit)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer cancel()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
	}()
	l.Infof("example-agent started")

	<-ctxExit.Done()
	currentApp.Stop()

	l.Infof("example-agent shutting down")
}
