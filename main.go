package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/a-h/templ"
	"github.com/labstack/echo/v5"
	"github.com/zmaillard/whereami/config"
	"github.com/zmaillard/whereami/db"
	"github.com/zmaillard/whereami/handlers"
	"github.com/zmaillard/whereami/middleware"
)

var Version string // Injected by ldflags at build time

func init() {
	db.RegisterExtensions()
}

func NewSecureClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: 5 * time.Second,
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.NewConfigWithVersion(Version)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		panic(err)
	}

	database, err := db.NewDatabase(cfg)
	if err != nil {
		logger.Error("failed to init database", "error", err)
		panic(err)
	}

	httpClient := NewSecureClient()

	e := echo.New()

	logCtx := middleware.WithLogger(context.Background(), logger)
	ctx, stop := signal.NotifyContext(logCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:         ":8080",
		GracefulTimeout: 5 * time.Second,
	}

	err = database.LoadSummitTree()
	if err != nil {
		logger.Error("failed to load summit tree", "error", err)
	}
	e.Use(middleware.RequestLogger(logger))

	e.Static("/assets", "assets")

	e.GET("/", handlers.Index(cfg))
	e.POST("/", handlers.InitialQuery(database))
	e.POST("/query", handlers.Query(database, httpClient))

	//e.RouteNotFound("/*", handlers.NotFound(logger))

	logger.Info("server starting", "port", 8080)
	if err := sc.Start(ctx, e); err != nil {
		logger.Error("server failed to start", "error", err)
	}
}
