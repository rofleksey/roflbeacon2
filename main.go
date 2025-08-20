//nolint:gocritic
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"roflbeacon2/app/api"
	"roflbeacon2/app/controller"
	"roflbeacon2/app/service/account"
	"roflbeacon2/app/service/alert"
	"roflbeacon2/app/service/ingest"
	"roflbeacon2/app/service/limits"
	"roflbeacon2/app/service/offline"
	"roflbeacon2/app/service/telegram"
	"roflbeacon2/pkg/config"
	"roflbeacon2/pkg/database"
	"roflbeacon2/pkg/middleware"
	"roflbeacon2/pkg/migration"
	"roflbeacon2/pkg/routes"
	"roflbeacon2/pkg/tlog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/do"
	_ "go.uber.org/automaxprocs"
)

func main() {
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exitChan := make(chan struct{})

	di := do.New()
	do.ProvideValue(di, appCtx)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	do.ProvideValue(di, cfg)

	if err = tlog.Init(cfg); err != nil {
		log.Fatalf("logging init failed: %v", err)
	}
	slog.ErrorContext(appCtx, "Service restarted")

	dbConnStr := "postgres://" + cfg.DB.User + ":" + cfg.DB.Pass + "@" + cfg.DB.Host + "/" + cfg.DB.Database + "?sslmode=disable&pool_max_conns=30&pool_min_conns=5&pool_max_conn_lifetime=1h&pool_max_conn_idle_time=30m&pool_health_check_period=1m&connect_timeout=10"

	dbConf, err := pgxpool.ParseConfig(dbConnStr)
	if err != nil {
		log.Fatalf("pgxpool.ParseConfig() failed: %v", err)
	}

	dbConf.ConnConfig.RuntimeParams = map[string]string{
		"statement_timeout":                   "30000",
		"idle_in_transaction_session_timeout": "60000",
	}

	dbConn, err := pgxpool.NewWithConfig(appCtx, dbConf)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	if err = database.InitSchema(appCtx, dbConn); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	do.ProvideValue(di, dbConn)

	queries := database.New(dbConn)
	do.ProvideValue(di, queries)

	if err = migration.Migrate(appCtx, di); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	do.Provide(di, account.New)
	do.Provide(di, telegram.New)
	do.Provide(di, alert.New)
	do.Provide(di, limits.New)
	do.Provide(di, ingest.New)
	do.Provide(di, offline.New)

	go do.MustInvoke[*telegram.Service](di).Run(appCtx)
	go do.MustInvoke[*offline.Service](di).RunBackgroundChecks(appCtx)

	server := controller.NewStrictServer(di)
	handler := api.NewStrictHandler(server, nil)

	app := fiber.New(fiber.Config{
		AppName:          "RoflBeacon2 API",
		BodyLimit:        1024 * 1024 * 10, // 10MB
		ErrorHandler:     middleware.ErrorHandler,
		ProxyHeader:      "X-Forwarded-For",
		ReadTimeout:      time.Second * 60,
		WriteTimeout:     time.Second * 60,
		DisableKeepalive: false,
	})

	middleware.FiberMiddleware(app, di)

	apiGroup := app.Group("/v1")
	api.RegisterHandlersWithOptions(apiGroup, handler, api.FiberServerOptions{
		BaseURL: "",
		Middlewares: []api.MiddlewareFunc{
			middleware.NewOpenAPIValidator(),
		},
	})

	routes.NotFoundRoute(app)

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Info("Shutting down server...")

		if err := app.Shutdown(); err != nil {
			log.Infof("Server is shutting down! Reason: %v", err)
		}

		close(exitChan)
	}()

	log.Info("Server started on port 8080")

	go func() {
		http.Handle("/metrics", promhttp.Handler())

		log.Info("Started metrics server on port 8081")
		if err := http.ListenAndServe(":8081", nil); err != nil { //nolint:gosec
			log.Warnf("failed to start metrics server: %v", err)
		}
	}()

	if err := app.Listen(":8080"); err != nil {
		log.Infof("Server stopped! Reason: %v", err)
	}

	<-exitChan
	cancel()

	log.Info("Waiting for services to finish...")
	_ = di.Shutdown()
}
