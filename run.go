package aruna

import (
	"context"

	"github.com/rlanhellas/aruna/config"
	"github.com/rlanhellas/aruna/httpbridge"
)

// RunRequest contains all configuration to run your app
type RunRequest struct {
	RoutesGroup    []*httpbridge.RouteGroupHttp
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
		go req.BackgroundTask(ctx)
	}

	if config.HttpServerEnabled() {
		go setupHttpServer(req.RoutesGroup, ctx)
	}

	if config.DbEnabled() {
		setupDB(ctx, req.MigrateTables)
	}

	//setupAuthZAuthN()

	//run forever
	for {
		select {}
	}
}
