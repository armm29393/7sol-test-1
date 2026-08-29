package user

import (
	"errors"
	"net/http"

	userdomain "user-management/internal/domain/user"
	userusecase "user-management/internal/usecase/user"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	uc *userusecase.Usecase
}

func NewHandler(uc *userusecase.Usecase) *Handler { return &Handler{uc: uc} }

// RegisterRoutes mounts every route owned by the user module. The auth
// middleware is passed in so this package stays independent of the router.
func RegisterRoutes(e *echo.Echo, h *Handler, auth echo.MiddlewareFunc) {
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)

	users := e.Group("/users", auth)
	users.GET("", h.List)
	users.GET("/:id", h.GetByID)
	users.PUT("/:id", h.Update)
	users.DELETE("/:id", h.Delete)
}

// errorResponse maps domain errors onto HTTP status codes. Unknown errors
// become a generic 500 so internal details never reach the client.
func errorResponse(err error) error {
	switch {
	case errors.Is(err, userdomain.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, userdomain.ErrEmailExists):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, userdomain.ErrInvalidCreds):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}

func bindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (h *Handler) Register(c echo.Context) error {
	var req userdomain.RegisterReq
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.uc.Register(c.Request().Context(), req.Name, req.Email, req.Password)
	if err != nil {
		return errorResponse(err)
	}
	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) Login(c echo.Context) error {
	var req userdomain.LoginReq
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	token, err := h.uc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return errorResponse(err)
	}
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

func (h *Handler) GetByID(c echo.Context) error {
	u, err := h.uc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return errorResponse(err)
	}
	return c.JSON(http.StatusOK, u)
}

func (h *Handler) List(c echo.Context) error {
	users, err := h.uc.List(c.Request().Context())
	if err != nil {
		return errorResponse(err)
	}
	return c.JSON(http.StatusOK, users)
}

func (h *Handler) Update(c echo.Context) error {
	var req userdomain.UpdateReq
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.uc.Update(c.Request().Context(), c.Param("id"), req.Name, req.Email)
	if err != nil {
		return errorResponse(err)
	}
	return c.JSON(http.StatusOK, u)
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.uc.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return errorResponse(err)
	}
	return c.NoContent(http.StatusNoContent)
}
