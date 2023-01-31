package aruna

import (
	"context"

	"github.com/rlanhellas/aruna/config"
	"github.com/rlanhellas/aruna/httpbridge"
	"github.com/rlanhellas/aruna/logger"
)

// RunRequest contains all configuration to run your app
type RunRequest struct {
	RoutesHttp     []*httpbridge.RouteHttp
	MigrateTables  []any
	BackgroundTask func(ctx context.Context)
}

// Run Starts the application
func Run(req *RunRequest) {
	ctx := context.Background()
	setupConfig()
	setupLogger()

	//go setupMetrics()

	if req.BackgroundTask != nil {
		logger.Info(ctx, "Running background task ...")
		go req.BackgroundTask(ctx)
	}

	if config.HttpServerEnabled() {
		logger.Info(ctx, "Running HTTP server ...")
		go setupHttpServer(req.RoutesHttp, ctx)
	}

	if config.DbEnabled() {
		logger.Info(ctx, "Configuring Database configuration ...")
		setupDB(ctx, req.MigrateTables)
	}

	//setupAuthZAuthN()

	//run forever
	for {
		select {}
	}
}
