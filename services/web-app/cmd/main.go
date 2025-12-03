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
	grpcMiddleware "github.com/yourorg/boilerplate/shared/middleware/grpc"
	httpMiddleware "github.com/yourorg/boilerplate/shared/middleware/http"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	"github.com/yourorg/boilerplate/shared/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TemplateRenderer is a custom html/template renderer for Echo
type TemplateRenderer struct {
	templates map[string]*template.Template
	partials  map[string]*template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Check if it's a partial (starts with "partials/")
	if len(name) > 9 && name[:9] == "partials/" {
		tmpl, ok := t.partials[name]
		if !ok {
			return fmt.Errorf("partial template %s not found", name)
		}
		// Get the base filename for ExecuteTemplate (e.g., "user-stats.html")
		baseName := name[len("partials/"):]
		return tmpl.ExecuteTemplate(w, baseName, data)
	}

	// Regular page template with base layout
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
	log := logger.NewWithService(cfg.LogLevel, cfg.LogFormat, "web-app")
	log.Info().Msg("Starting web-app service")

	// Initialize tracing
	ctx := context.Background()
	tp, err := tracing.InitTracer(ctx, tracing.Config{
		ServiceName:    "web-app",
		JaegerEndpoint: cfg.OTLPEndpoint,
		Environment:    cfg.Environment,
		Enabled:        cfg.TracingEnabled,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize tracing, continuing without it")
	} else if tp != nil {
		defer tracing.Shutdown(ctx, tp)
		log.Info().Str("endpoint", cfg.OTLPEndpoint).Msg("Tracing initialized")
	} else {
		log.Info().Msg("Tracing disabled")
	}

	// Initialize template renderer with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"formatTime": func(t interface{}) string {
			switch v := t.(type) {
			case *timestamppb.Timestamp:
				if v == nil {
					return "N/A"
				}
				return v.AsTime().Format("Jan 02, 2006 15:04")
			case time.Time:
				return v.Format("Jan 02, 2006 15:04")
			default:
				return fmt.Sprintf("%v", t)
			}
		},
	}

	// Parse templates - each page template is combined with the base layout
	templates := make(map[string]*template.Template)
	partials := make(map[string]*template.Template)

	// Parse all page templates
	pages := []string{
		"templates/pages/login.html",
		"templates/pages/signup.html",
		"templates/pages/dashboard.html",
	}

	for _, page := range pages {
		name := page[len("templates/pages/"):]  // Extract filename: "login.html", "dashboard.html"
		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(
			"templates/layouts/base.html",
			page,
		))
		templates[name] = tmpl
	}

	// Parse partial templates (for HTMX responses)
	partialFiles := []string{
		"templates/partials/user-stats.html",
		"templates/partials/recent-activity.html",
		"templates/partials/user-list.html",
	}

	for _, partial := range partialFiles {
		name := partial[len("templates/"):]  // Extract: "partials/user-stats.html"
		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(partial))
		partials[name] = tmpl
	}

	renderer := &TemplateRenderer{
		templates: templates,
		partials:  partials,
	}

	// Initialize Echo
	e := echo.New()
	e.Renderer = renderer
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(httpMiddleware.TracingMiddlewareWithConfig(httpMiddleware.TracingConfig{
		SkipPaths: []string{"/metrics", "/health", "/static"},
	}))
	e.Use(httpMiddleware.MetricsMiddlewareWithConfig(httpMiddleware.MetricsConfig{
		SkipPaths: []string{"/metrics", "/health", "/static"},
	}))
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Metrics endpoint
	e.GET("/metrics", httpMiddleware.MetricsHandler())

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

	// Initialize JWT-based session store
	sessionStore := session.NewStore(cfg.JWTSecret, cfg.JWTExpiry)

	// Connect to user service
	log.Info().Str("address", cfg.UserServiceAddr).Msg("Connecting to user service")
	userConn, err := grpc.NewClient(
		cfg.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcMiddleware.ClientTracingInterceptor()),
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
	e.GET("/signup", authHandler.ShowSignupPage, mw.RedirectIfAuthenticated(sessionStore))
	e.POST("/signup", authHandler.Signup)
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
