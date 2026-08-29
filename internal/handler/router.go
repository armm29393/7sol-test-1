package handler

import (
	"github.com/labstack/echo/v4"

	userhandler "user-management/internal/handler/user"
)

// NewRouter builds the echo instance with the validator and shared middleware,
// then lets each module register its own routes.
func NewRouter(uh *userhandler.Handler, jwtSecret string) *echo.Echo {
	e := echo.New()
	e.Validator = NewValidator()
	e.Use(RequestLogger())

	userhandler.RegisterRoutes(e, uh, JWTMiddleware(jwtSecret))

	return e
}
