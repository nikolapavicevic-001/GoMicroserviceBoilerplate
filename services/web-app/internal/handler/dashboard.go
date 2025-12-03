package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yourorg/boilerplate/services/web-app/internal/session"
	commonv1 "github.com/yourorg/boilerplate/shared/proto/gen/common/v1"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
)

type DashboardHandler struct {
	sessionStore *session.Store
	userClient   pb.UserServiceClient
}

func NewDashboardHandler(sessionStore *session.Store, userClient pb.UserServiceClient) *DashboardHandler {
	return &DashboardHandler{
		sessionStore: sessionStore,
		userClient:   userClient,
	}
}

// ShowDashboard renders the main dashboard page
func (h *DashboardHandler) ShowDashboard(c echo.Context) error {
	userID, email, name, ok := h.sessionStore.GetUser(c)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"Title":  "Dashboard",
		"UserID": userID,
		"Email":  email,
		"Name":   name,
	})
}

// GetUserStats returns user statistics (HTMX partial)
func (h *DashboardHandler) GetUserStats(c echo.Context) error {
	userID, _, _, ok := h.sessionStore.GetUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	// Get user details
	resp, err := h.userClient.GetUser(ctx, &pb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load user stats")
	}

	user := resp.User

	return c.Render(http.StatusOK, "partials/user-stats.html", map[string]interface{}{
		"User":      user,
		"CreatedAt": user.CreatedAt,
		"UpdatedAt": user.UpdatedAt,
	})
}

// GetRecentActivity returns recent activity (HTMX partial)
func (h *DashboardHandler) GetRecentActivity(c echo.Context) error {
	userID, _, _, ok := h.sessionStore.GetUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	// In a real app, you would fetch actual activity data
	// For demo purposes, we'll return mock data
	activities := []map[string]interface{}{
		{
			"Action":    "Logged in",
			"Timestamp": time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
		},
		{
			"Action":    "Updated profile",
			"Timestamp": time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05"),
		},
		{
			"Action":    "Changed password",
			"Timestamp": time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05"),
		},
	}

	return c.Render(http.StatusOK, "partials/recent-activity.html", map[string]interface{}{
		"Activities": activities,
		"UserID":     userID,
	})
}

// GetUserList returns a list of users (HTMX partial)
func (h *DashboardHandler) GetUserList(c echo.Context) error {
	_, _, _, ok := h.sessionStore.GetUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	// List users
	resp, err := h.userClient.ListUsers(ctx, &pb.ListUsersRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:     1,
			PageSize: 10,
		},
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load users")
	}

	return c.Render(http.StatusOK, "partials/user-list.html", map[string]interface{}{
		"Users": resp.Users,
		"Total": resp.Pagination.TotalCount,
	})
}
