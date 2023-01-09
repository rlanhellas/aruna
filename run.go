package aruna

import (
	"context"
	"github.com/rlanhellas/aruna/config"
	"github.com/rlanhellas/aruna/httpbridge"
)

// RunRequest contains all configuration to run your app
type RunRequest struct {
	RoutesHttp    []*httpbridge.RouteHttp
	MigrateTables []any
}

// Run Starts the application
func Run(req *RunRequest) {
	ctx := context.Background()
	setupConfig()
	setupLogger()

	//go setupMetrics()

	if config.HttpServerEnabled() {
		go setupHttpServer(req.RoutesHttp, ctx)
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
