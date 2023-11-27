package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"errors"
	"net/http"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/app/services/sales-api/handlers"
	database "github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/sys/database/pgx"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/auth"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/v1/debug"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/keystore"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/logger"
	"github.com/ardanlabs/conf/v3"
	"go.uber.org/zap"
)

var build = "develop"

func main() {
	log, err := logger.New("SALES-API")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	ctx := context.Background()

	if err := run(log, ctx); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

/*
	//TODO:
	Need to figure out timeouts for http service.
	Add Category field and type to product.
*/

func run(log *zap.SugaredLogger, ctx context.Context) error {
	// -----------------------------------------------------------------------
	// GOMAXPROCS
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0), "BUILD-", build)

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:3000,mask"` //mask print it as xxxxxx
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:database-service.sales-system.svc.cluster.local"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:2"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		Auth struct {
			KeysFolder string `conf:"default:zarf/keys/"`
			ActiveKID  string `conf:"default:private"`
			Issuer     string `conf:"default:service project"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "BILL KENNEDY",
		},
	}

	const prefix = "SALES"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Info("starting service ", "version ", build)
	defer log.Info("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}

	log.Info("startup", "config", out)

	// -------------------------------------------------------------------------
	// Database Support

	log.Info(ctx, "startup", "status", "initializing database support", "host", cfg.DB.Host)

	db, err := database.Open(database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Info(ctx, "shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()

	// Simple keystore.
	ks, err := keystore.NewFS(os.DirFS(cfg.Auth.KeysFolder))
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	authCfg := auth.Config{
		Log:       log,
		KeyLookup: ks,
		Issuer:    cfg.Auth.Issuer,
	}
	fmt.Println(authCfg)
	auth, err := auth.New(authCfg)
	if err != nil {
		return fmt.Errorf("constructing auth: %w", err)
	}

	// -------------------------------------------------------------------------
	// Start Debug Service
	// This creats a go that blocks on a listening serve call on whatever the IP for the debug host is

	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux(build, log, db)); err != nil {
			log.Error("shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -----------------------------------------------------------------------
	// Start API Service
	log.Infow("startup", "status", "Initializing V1 API support")

	serverErrors := make(chan error, 1)

	// -----------------------------------------------------------------------
	shutdown := make(chan os.Signal, 1)

	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
		Auth:     auth,
		DB:       db,
	})

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()
	// we are waiting for SIGINT which is a Ctrl+C or
	// a SIGTERM which what will get back from Kubernetes
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			// If we timeo
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
