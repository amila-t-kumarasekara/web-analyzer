package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/config"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/fetcher"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/httpclient"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/linkcheck"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/web"
	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.LoadConfig()

	fetchClient := fetcher.New(httpclient.NewPageClient(), consts.DefaultMaxPageBytes)
	linkChecker := linkcheck.New(httpclient.NewLinkClient(), consts.DefaultLinkWorkers)
	handler := web.NewHandler(fetchClient, linkChecker)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: consts.ReadHeaderTimeout,
		ReadTimeout:       consts.ReadTimeout,
		WriteTimeout:      consts.WriteTimeout,
		IdleTimeout:       consts.IdleTimeout,
	}

	var adminSrv *http.Server
	if cfg.AdminAddr != "" {
		adminMux := http.NewServeMux()
		adminMux.HandleFunc("/debug/pprof/", pprof.Index)
		adminMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		adminMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		adminMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		adminMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		adminMux.Handle("/metrics", promhttp.Handler())
		adminSrv = &http.Server{
			Addr:              cfg.AdminAddr,
			Handler:           adminMux,
			ReadHeaderTimeout: consts.ReadHeaderTimeout,
			ReadTimeout:       consts.ReadTimeout,
			WriteTimeout:      consts.WriteTimeout,
			IdleTimeout:       consts.IdleTimeout,
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", slog.String("port", cfg.Port))
		serveErr <- srv.ListenAndServe()
	}()

	if adminSrv != nil {
		go func() {
			logger.Info("admin server listening", slog.String("addr", cfg.AdminAddr))
			if err := adminSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("admin server", slog.Any("err", err))
			}
		}()
	}

	select {
	case <-ctx.Done():
	case err := <-serveErr:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server", slog.Any("err", err))
			os.Exit(1)
		}
	}
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), consts.ShutdownTimeout)
	defer cancel()

	if adminSrv != nil {
		if err := adminSrv.Shutdown(shutdownCtx); err != nil {
			logger.Error("admin shutdown", slog.Any("err", err))
		}
	}
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", slog.Any("err", err))
		os.Exit(1)
	}
}
