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
	echoprometheus "github.com/labstack/echo-prometheus"
	"github.com/labstack/echo/v5"
	"github.com/zmaillard/whereami/config"
	"github.com/zmaillard/whereami/handlers"
	"github.com/zmaillard/whereami/metrics"
	"github.com/zmaillard/whereami/middleware"
	"github.com/zmaillard/whereami/queries"
)

var Version string // Injected by ldflags at build time

func init() {
	queries.RegisterExtensions()
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
		Timeout: 10 * time.Second,
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	cfg, err := config.NewConfigWithVersion(Version)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := queries.NewDatabase(cfg)
	if err != nil {
		logger.Error("failed to init database", "error", err)
		panic(err)
	}

	err = database.LoadSummitTree(ctx)
	if err != nil {
		logger.Error("failed to init build summit index", "error", err)
		panic(err)
	}

	httpClient := NewSecureClient()

	// Initialize custom Prometheus metrics
	metrics.Init()

	e := echo.New()
	e.Use(echoprometheus.NewMiddleware("whereami"))
	go func() {
		metrics := echo.New()                                // this Echo will run on separate port 8081
		metrics.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics
		if err := metrics.Start(":8081"); err != nil {
			e.Logger.Error("failed to start metrics server", "error", err)
		}
	}()

	sc := echo.StartConfig{
		Address:         "0.0.0.0:8080",
		GracefulTimeout: 5 * time.Second,
	}

	e.Use(middleware.RequestLogger(logger))

	e.Static("/assets", "assets")

	e.GET("/", handlers.Index(cfg))
	e.GET("/about", handlers.About(cfg))
	e.GET("/map", handlers.Map(cfg))
	e.POST("/details", handlers.Details(database, httpClient, cfg))
	e.POST("/geocode", handlers.Geocode(database))

	//e.RouteNotFound("/*", handlers.NotFound(logger))

	logger.Info("server starting", "port", 8080)
	if err := sc.Start(ctx, e); err != nil {
		logger.Error("server failed to start", "error", err)
	}
}
