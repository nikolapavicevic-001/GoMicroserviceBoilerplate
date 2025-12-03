package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yourorg/boilerplate/services/web-app/config"
	"github.com/yourorg/boilerplate/services/web-app/internal/handler"
	mw "github.com/yourorg/boilerplate/services/web-app/internal/middleware"
	"github.com/yourorg/boilerplate/services/web-app/internal/session"
	"github.com/yourorg/boilerplate/shared/auth"
	"github.com/yourorg/boilerplate/shared/logger"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TemplateRenderer is a custom html/template renderer for Echo
type TemplateRenderer struct {
	templates map[string]*template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	log.Info().Msg("Starting web-app service")

	// Initialize template renderer with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Parse templates - each page template is combined with the base layout
	templates := make(map[string]*template.Template)

	// Parse all page templates
	pages := []string{
		"templates/pages/login.html",
		"templates/pages/dashboard.html",
	}

	for _, page := range pages {
		name := page[len("templates/pages/"):]  // Extract filename: "login.html", "dashboard.html"
		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(
			"templates/layouts/base.html",
			page,
			"templates/partials/user-stats.html",
			"templates/partials/recent-activity.html",
			"templates/partials/user-list.html",
		))
		templates[name] = tmpl
	}

	renderer := &TemplateRenderer{
		templates: templates,
	}

	// Initialize Echo
	e := echo.New()
	e.Renderer = renderer
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Logger middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Log request
			log.Info().
				Str("method", c.Request().Method).
				Str("uri", c.Request().RequestURI).
				Int("status", c.Response().Status).
				Dur("latency", time.Since(start)).
				Str("ip", c.RealIP()).
				Msg("HTTP request")

			return err
		}
	})

	// Static files
	e.Static("/static", "static")

	// Initialize session store
	sessionStore := session.NewStore(cfg.SessionSecret, cfg.SessionMaxAge)

	// Connect to user service
	log.Info().Str("address", cfg.UserServiceAddr).Msg("Connecting to user service")
	userConn, err := grpc.NewClient(
		cfg.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to user service")
	}
	defer userConn.Close()
	userClient := pb.NewUserServiceClient(userConn)

	// Initialize OAuth2 providers
	var googleProvider, githubProvider *auth.OAuth2Provider
	if cfg.OAuth2GoogleClientID != "" {
		googleProvider = auth.NewGoogleProvider(
			cfg.OAuth2GoogleClientID,
			cfg.OAuth2GoogleClientSecret,
			cfg.OAuth2GoogleRedirectURL,
		)
		log.Info().Msg("Google OAuth2 provider initialized")
	}
	if cfg.OAuth2GitHubClientID != "" {
		githubProvider = auth.NewGitHubProvider(
			cfg.OAuth2GitHubClientID,
			cfg.OAuth2GitHubClientSecret,
			cfg.OAuth2GitHubRedirectURL,
		)
		log.Info().Msg("GitHub OAuth2 provider initialized")
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(sessionStore, userClient, googleProvider, githubProvider)
	dashboardHandler := handler.NewDashboardHandler(sessionStore, userClient)

	// Routes - Public
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/login")
	})
	e.GET("/login", authHandler.ShowLoginPage, mw.RedirectIfAuthenticated(sessionStore))
	e.POST("/login", authHandler.Login)
	e.GET("/logout", authHandler.Logout)

	// OAuth2 routes
	e.GET("/auth/google", authHandler.GoogleLogin)
	e.GET("/auth/google/callback", authHandler.GoogleCallback)
	e.GET("/auth/github", authHandler.GitHubLogin)
	e.GET("/auth/github/callback", authHandler.GitHubCallback)

	// Routes - Protected
	protected := e.Group("")
	protected.Use(mw.RequireAuth(sessionStore))

	protected.GET("/dashboard", dashboardHandler.ShowDashboard)
	protected.GET("/dashboard/stats", dashboardHandler.GetUserStats)
	protected.GET("/dashboard/activity", dashboardHandler.GetRecentActivity)
	protected.GET("/dashboard/users", dashboardHandler.GetUserList)

	// Demo endpoint for live updates
	protected.GET("/dashboard/time", func(c echo.Context) error {
		return c.String(http.StatusOK, time.Now().Format("15:04:05"))
	})

	// Demo endpoint for HTMX action
	protected.POST("/dashboard/action", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<div class="mt-2 text-sm text-green-600">Action completed successfully!</div>`)
	})

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "web-app",
		})
	})

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", cfg.HTTPPort)
		log.Info().Str("address", addr).Msg("Starting HTTP server")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
