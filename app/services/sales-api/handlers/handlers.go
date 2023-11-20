package handlers

import (
	"net/http"
	"os"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/app/services/sales-api/handlers/v1/testgrp"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/v1/mid"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) *web.App {
	app := web.NewApp(cfg.Shutdown, mid.Logger(cfg.Log), mid.Errors(cfg.Log), mid.Panics())

	app.Handle(http.MethodGet, "/test", testgrp.Test)
	return app
}
