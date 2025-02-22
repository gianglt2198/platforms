package app

import (
	"fmt"

	"github.com/ansrivas/fiberprometheus/v2"
	oblogger "github.com/gianglt2198/platforms/observability/logger"
	"github.com/gianglt2198/platforms/services/rest/config"
	"github.com/gianglt2198/platforms/services/rest/middlewares"
	"github.com/gianglt2198/platforms/services/rest/routes"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
)

type App struct {
	cfg      *config.Config
	db       *gorm.DB
	app      *fiber.App
	logger   oblogger.ObLogger
	handlers []Handler
}

type Handler interface {
	Register(router fiber.Router)
}

func New(cfg *config.Config, db *gorm.DB, logger oblogger.ObLogger) *App {

	app := fiber.New()

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(middlewares.RequestIDMiddleware)
	app.Use(middlewares.TracingMiddleware("main", "request_caller",
		middlewares.TracingConfig{
			ServiceName:    cfg.App.Name,
			ServiceVersion: "1.0.0",
		}))
	app.Use(middlewares.MetricMiddleware(middlewares.MetricConfig{
		ServiceName:    cfg.App.Name,
		ServiceVersion: "1.0.0",
	}))

	prometheus := fiberprometheus.New(cfg.App.Name)
	prometheus.RegisterAt(app, "/metrics")
	prometheus.SetSkipPaths([]string{
		"/metrics", "/health", "/swagger",
	})

	app.Get("/health", HealthCheck)

	return &App{
		cfg:    cfg,
		db:     db,
		app:    app,
		logger: logger,
	}
}

func (a *App) RegisterHandlers(handlers ...Handler) {
	a.handlers = handlers

	api := a.app.Group("/api")
	for _, h := range a.handlers {
		h.Register(api)
	}

	a.app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(
			routes.ErrorResponse(fiber.ErrBadGateway),
		)
	})
}

func (a *App) RegisterSwagger(fileContent []byte) {
	a.app.Use(swagger.New(swagger.Config{
		Title:       "Swagger API",
		FileContent: fileContent,
	}))
}

func (a *App) Start() error {
	return a.app.Listen(fmt.Sprintf(":%d", a.cfg.App.Port))
}

func HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
}
