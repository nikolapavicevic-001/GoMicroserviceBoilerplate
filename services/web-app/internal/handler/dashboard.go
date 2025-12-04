package handler

import (

	"github.com/labstack/echo/v4"
	"github.com/yourorg/boilerplate/shared/auth"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
)

const (
	CookieName = "auth_token"
)

type DashboardHandler struct {
	jwtSecret  string
	userClient pb.UserServiceClient
}

func NewDashboardHandler(jwtSecret string, userClient pb.UserServiceClient) *DashboardHandler {
	return &DashboardHandler{
		jwtSecret:  jwtSecret,
		userClient: userClient,
	}
}

// getUserFromCookie extracts user info from JWT cookie
func (h *DashboardHandler) getUserFromCookie(c echo.Context) (userID, email string, ok bool) {
	cookie, err := c.Cookie(CookieName)
	if err != nil {
		return "", "", false
	}

	claims, err := auth.ValidateToken(cookie.Value, h.jwtSecret)
	if err != nil {
		return "", "", false
	}

	return claims.UserID, claims.Email, true
}

func (h *DashboardHandler) ShowDashboard(c echo.Context) error {
	return c.Render(200, "dashboard.html", map[string]interface{}{
	})
}

