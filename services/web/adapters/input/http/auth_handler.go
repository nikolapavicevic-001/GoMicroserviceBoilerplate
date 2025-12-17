package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/microserviceboilerplate/web/domain/entity"
	"github.com/microserviceboilerplate/web/domain/port/input"
	"github.com/microserviceboilerplate/web/domain/port/output"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authUseCase  input.AuthUseCase
	templateRepo output.TemplateRepository
	sessionRepo  output.SessionRepository
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authUseCase input.AuthUseCase,
	templateRepo output.TemplateRepository,
	sessionRepo output.SessionRepository,
) *AuthHandler {
	return &AuthHandler{
		authUseCase:  authUseCase,
		templateRepo: templateRepo,
		sessionRepo:  sessionRepo,
	}
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/auth/login", h.LoginPage)
	r.Post("/auth/login", h.Login)
	r.Get("/auth/register", h.RegisterPage)
	r.Post("/auth/register", h.Register)
	r.Post("/auth/logout", h.Logout)
}

// LoginPage renders the login page
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Login",
	}

	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	html, err := h.templateRepo.RenderString(ctx, "login.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// Login handles login requests
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	// Avoid common copy/paste issues (e.g. trailing newline) while still not logging secrets.
	password := strings.TrimSpace(r.FormValue("password"))

	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	req := input.LoginRequest{
		Email:    email,
		Password: password,
	}

	user, err := h.authUseCase.Login(ctx, req)
	if err != nil {
		log.Printf("login failed for %q: %v", email, err)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Retarget", "#error-message")
			w.Header().Set("HX-Reswap", "innerHTML")
			fmt.Fprintf(w, `<div class="error">Invalid credentials</div>`)
		} else {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
		return
	}

	// Save session
	userEntity := entity.NewUser(user.ID, user.Email, user.Username, user.Name)
	userEntity.IsAuthenticated = true

	if err := h.sessionRepo.Save(ctx, "", userEntity); err != nil {
		log.Printf("Error saving session: %v", err)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Retarget", "#error-message")
			w.Header().Set("HX-Reswap", "innerHTML")
			fmt.Fprintf(w, `<div class="error">Failed to save session: %v</div>`, err)
		} else {
			http.Error(w, "Failed to save session", http.StatusInternalServerError)
		}
		return
	}

	// Handle redirect
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<script>window.location.replace("/dashboard");</script>`)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}

// RegisterPage renders the registration page
func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Register",
	}

	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	html, err := h.templateRepo.RenderString(ctx, "register.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// Register handles registration requests
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	password := strings.TrimSpace(r.FormValue("password"))
	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")

	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	req := input.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	user, err := h.authUseCase.Register(ctx, req)
	if err != nil {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Retarget", "#error-message")
			w.Header().Set("HX-Reswap", "innerHTML")
			fmt.Fprintf(w, `<div class="error">Registration failed. User may already exist.</div>`)
		} else {
			http.Error(w, "Registration failed", http.StatusBadRequest)
		}
		return
	}

	// Save session
	userEntity := entity.NewUser(user.ID, user.Email, user.Username, user.Name)
	userEntity.IsAuthenticated = true

	if err := h.sessionRepo.Save(ctx, "", userEntity); err != nil {
		log.Printf("Error saving session: %v", err)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Retarget", "#error-message")
			w.Header().Set("HX-Reswap", "innerHTML")
			fmt.Fprintf(w, `<div class="error">Failed to save session: %v</div>`, err)
		} else {
			http.Error(w, "Failed to save session", http.StatusInternalServerError)
		}
		return
	}

	// Handle redirect
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<script>window.location.replace("/dashboard");</script>`)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}

// Logout handles logout requests
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	if err := h.authUseCase.Logout(ctx, ""); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/auth/login")
	} else {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
	}
}
