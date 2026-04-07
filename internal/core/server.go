package core

import (
	"log"

	"github.com/parxyws/cozybox/internal/handlers"
	"github.com/parxyws/cozybox/internal/middleware"
	"github.com/parxyws/cozybox/internal/routes"
	"github.com/parxyws/cozybox/internal/service"
	"github.com/parxyws/cozybox/internal/tools/mail"
	"github.com/parxyws/cozybox/internal/tools/util"
)

func (a *App) Bootstrap() error {
	// ── JWT ─────────────────────────────────────────────────────────────
	jwt, err := util.NewJWTMaker(a.cfg.Server.JWTSecretKey)
	if err != nil {
		return err
	}

	// ── Middleware ───────────────────────────────────────────────────────
	mw := middleware.NewMiddlewareManager(&middleware.ConfigMiddleware{
		Logger: a.logger,
		Config: a.cfg,
		DB:     a.db,
	})

	// Global middleware
	a.app.Use(mw.CORSMiddleware())
	a.app.Use(mw.RequestIDMiddleware())

	// ── Services ────────────────────────────────────────────────────────
	mailer := mail.NewMailer(a.mail)

	userService := service.NewUserService(a.db, mailer, a.authRedis, jwt)

	// ── Handlers ────────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)

	// ── Routes ──────────────────────────────────────────────────────────
	api := a.app.Group("/api")

	// Public routes (no auth required)
	routes.AuthRoute(api, authHandler)

	// Protected routes (auth + tenant scope required)
	protected := api.Group("")
	protected.Use(mw.AuthMiddleware(jwt))
	protected.Use(mw.TenantScopeMiddleware())

	routes.AuthProtectedRoute(protected, authHandler)
	routes.UserRoute(protected, userHandler)

	log.Printf("Routes registered. Server ready on %s:%d", a.cfg.Server.Host, a.cfg.Server.Port)

	return nil
}
