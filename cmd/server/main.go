package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"speaknow/internal/assets"
	"speaknow/internal/handler"
	"speaknow/internal/middleware"
	"speaknow/internal/provider/factory"
	"speaknow/internal/service/asr"
	"speaknow/internal/service/cache"
	"speaknow/internal/service/cost"
	"speaknow/internal/service/router"
	"speaknow/pkg/logger"
)

func main() {
	configPath := flag.String("config", "", "config file path (empty = use built-in default)")
	flag.Parse()

	if err := assets.Prepare(); err != nil {
		fmt.Fprintf(os.Stderr, "prepare assets: %v\n", err)
		os.Exit(1)
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	if cfg.Providers.Vosk.Enabled && (cfg.Providers.Vosk.ModelPath == "" || cfg.Providers.Vosk.ModelPath == "embedded") {
		cfg.Providers.Vosk.ModelPath = assets.ModelPath()
	}

	log, err := logger.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	registry, err := factory.BuildRegistry(cfg)
	if err != nil {
		log.Fatal("build provider registry", zap.Error(err))
	}

	cacheSvc := cache.NewService(cfg.Cache.TTL)
	routerSvc := router.New(registry, &cfg.ASR, log)
	costSvc := cost.New()
	asrSvc := asr.New(routerSvc, cacheSvc, costSvc, log)

	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit)

	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.UserID())
	r.Use(middleware.AccessLog(log))
	r.Use(rateLimiter.Middleware())
	r.Use(middleware.RetryAfterMiddleware())

	asrHandler := handler.NewASRHandler(asrSvc, routerSvc, costSvc, cfg.ASR.MaxAudioSize)
	wsHandler := handler.NewWSHandler(routerSvc)
	healthHandler := handler.NewHealthHandler(nil)

	webHandler, err := assets.WebHandler()
	if err != nil {
		log.Fatal("web handler", zap.Error(err))
	}
	r.GET("/web", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/")
	})
	r.GET("/web/*filepath", gin.WrapH(http.StripPrefix("/web", webHandler)))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/")
	})
	r.GET("/api/v1/health", healthHandler.Health)
	r.GET("/api/v1/providers/status", asrHandler.ProviderStatus)
	r.GET("/api/v1/stats/cost", asrHandler.CostStats)
	r.POST("/api/v1/asr/recognize", asrHandler.Recognize)
	r.GET("/api/v1/asr/stream", wsHandler.Stream)

	srvAddr := cfg.Server.Addr()
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	srv := &http.Server{
		Addr:              srvAddr,
		Handler:           r,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		BaseContext: func(net.Listener) context.Context {
			return rootCtx
		},
	}

	log.Info("SpeakNow server starting",
		zap.String("addr", srvAddr),
		zap.Strings("providers", registry.Names()),
		zap.Duration("read_header_timeout", cfg.Server.ReadHeaderTimeout),
		zap.Duration("read_timeout", cfg.Server.ReadTimeout),
		zap.Duration("write_timeout", cfg.Server.WriteTimeout),
		zap.Duration("idle_timeout", cfg.Server.IdleTimeout),
		zap.Duration("shutdown_timeout", cfg.Server.ShutdownTimeout),
	)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal("server stopped", zap.Error(err))
		}
	case sig := <-quit:
		log.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	rootCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown timed out or failed", zap.Error(err))
	} else {
		log.Info("server stopped gracefully")
	}

	signal.Stop(quit)
}
