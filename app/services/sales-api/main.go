package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"errors"
	"net/http"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/v1/debug"
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

	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
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
	// Start Debug Service
	// This creats a go that blocks on a listening serve call on whatever the IP for the debug host is

	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux()); err != nil {
			log.Error("shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -----------------------------------------------------------------------
	shutdown := make(chan os.Signal, 1)
	// we are waiting for SIGINT which is a Ctrl+C or
	// a SIGTERM which what will get back from Kubernetes
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdown

	log.Infow("shutdown", "status", "shutdown started", "singal", sig)
	defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

	return nil
}
