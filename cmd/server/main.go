package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"net/http"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"speaknow/internal/config"
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
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Warn("redis unavailable, cache and distributed rate limit disabled", zap.Error(err))
	}
	cancel()

	cacheSvc := cache.NewService(redisClient, cfg.Redis.CacheTTL)
	routerSvc := router.New(registry, &cfg.ASR, log)
	costSvc := cost.New()
	asrSvc := asr.New(routerSvc, cacheSvc, costSvc, log)

	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, redisClient)

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
	healthHandler := handler.NewHealthHandler(map[string]func(*gin.Context) error{
		"redis": func(c *gin.Context) error {
			return cacheSvc.Ping(c.Request.Context())
		},
	})

	r.Static("/web", "./web")
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/")
	})
	r.GET("/api/v1/health", healthHandler.Health)
	r.GET("/api/v1/providers/status", asrHandler.ProviderStatus)
	r.GET("/api/v1/stats/cost", asrHandler.CostStats)
	r.POST("/api/v1/asr/recognize", asrHandler.Recognize)
	r.GET("/api/v1/asr/stream", wsHandler.Stream)

	srvAddr := cfg.Server.Addr()
	log.Info("SpeakNow server starting",
		zap.String("addr", srvAddr),
		zap.Strings("providers", registry.Names()),
	)

	go func() {
		if err := r.Run(srvAddr); err != nil {
			log.Fatal("server stopped", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down...")
}
