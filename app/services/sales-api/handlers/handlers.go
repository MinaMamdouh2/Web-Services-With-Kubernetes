package handlers

import (
	"net/http"
	"os"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/app/services/sales-api/handlers/v1/testgrp"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/app/services/sales-api/handlers/v1/usrgrp"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/core/user"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/core/user/stores/userdb"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/auth"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/v1/mid"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	DB       *sqlx.DB
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) *web.App {
	app := web.NewApp(cfg.Shutdown, mid.Logger(cfg.Log), mid.Errors(cfg.Log), mid.Metrics(), mid.Panics())

	app.Handle(http.MethodGet, "/test", testgrp.Test)
	app.Handle(http.MethodGet, "/test/auth", testgrp.Test, mid.Authenticate(cfg.Auth), mid.Authorize(cfg.Auth, auth.RuleAdminOnly))

	// ==============================================================================
	usrCore := user.NewCore(userdb.NewStore(cfg.Log, cfg.DB))
	ugh := usrgrp.New(usrCore)
	app.Handle(http.MethodGet, "/users", ugh.Query)
	return app
}
