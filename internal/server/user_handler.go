package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"scaffold-api/internal/errs"
	"scaffold-api/internal/model"
	"scaffold-api/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// UserHandler handles user CRUD requests.
type UserHandler struct {
	service *service.UserService
}

// NewUserHandler creates a user handler.
func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// RegisterUserRoutes mounts user CRUD routes on the Echo engine.
func RegisterUserRoutes(e *echo.Echo, handler *UserHandler) {
	RegisterHealthRoutes(e)

	api := e.Group("/api/v1")
	users := api.Group("/users")
	users.POST("", handler.Create)
	users.GET("", handler.List)
	users.GET("/:id", handler.GetByID)
	users.PUT("/:id", handler.Update)
	users.PATCH("/:id", handler.Update)
	users.DELETE("/:id", handler.Delete)
}

// Create godoc
// @Summary Create user
// @Description Create a new user record.
// @Tags users
// @Accept json
// @Produce json
// @Param payload body service.CreateUserInput true "Create user payload"
// @Success 201 {object} UserDetailEnvelope
// @Failure 400 {object} ErrorEnvelope
// @Failure 409 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c echo.Context) error {
	var req service.CreateUserInput
	if err := c.Bind(&req); err != nil {
		return errs.NewValidation("invalid request body", map[string]string{
			"body": "request body is invalid",
		})
	}
	if err := c.Validate(&req); err != nil {
		return translateValidationError(err)
	}

	user, err := h.service.Create(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, UserDetailEnvelope{
		Data: newUserResponse(*user),
	})
}

// GetByID godoc
// @Summary Get user by ID
// @Description Fetch a single user by its numeric identifier.
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserDetailEnvelope
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c echo.Context) error {
	id, err := parseUserID(c.Param("id"))
	if err != nil {
		return err
	}

	user, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, UserDetailEnvelope{
		Data: newUserResponse(*user),
	})
}

// List godoc
// @Summary List users
// @Description List users with optional email and name filters.
// @Tags users
// @Produce json
// @Param email query string false "Filter by exact email"
// @Param name_like query string false "Filter by partial name match"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} UserListEnvelope
// @Failure 400 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/users [get]
func (h *UserHandler) List(c echo.Context) error {
	page, err := parseIntQuery(c, "page")
	if err != nil {
		return err
	}
	pageSize, err := parseIntQuery(c, "page_size")
	if err != nil {
		return err
	}

	input := service.ListUsersInput{
		Email:    optionalQuery(c, "email"),
		NameLike: optionalQuery(c, "name_like"),
		Page:     page,
		PageSize: pageSize,
	}

	result, err := h.service.List(c.Request().Context(), input)
	if err != nil {
		return err
	}

	items := make([]UserResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, newUserResponse(item))
	}

	return c.JSON(http.StatusOK, UserListEnvelope{
		Data: items,
		Pagination: Pagination{
			Page:     result.Page,
			PageSize: result.PageSize,
			Total:    result.Total,
		},
	})
}

// Update godoc
// @Summary Update user
// @Description Replace user fields by ID.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param payload body service.UpdateUserInput true "Update user payload"
// @Success 200 {object} UserDetailEnvelope
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 409 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/users/{id} [put]
// @Router /api/v1/users/{id} [patch]
func (h *UserHandler) Update(c echo.Context) error {
	id, err := parseUserID(c.Param("id"))
	if err != nil {
		return err
	}

	var req service.UpdateUserInput
	if err := c.Bind(&req); err != nil {
		return errs.NewValidation("invalid request body", map[string]string{
			"body": "request body is invalid",
		})
	}
	if err := c.Validate(&req); err != nil {
		return translateValidationError(err)
	}

	user, err := h.service.Update(c.Request().Context(), id, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, UserDetailEnvelope{
		Data: newUserResponse(*user),
	})
}

// Delete godoc
// @Summary Delete user
// @Description Delete a user by ID.
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 400 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c echo.Context) error {
	id, err := parseUserID(c.Param("id"))
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func newUserResponse(user model.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		UID:       user.UID,
		Email:     user.Email,
		Name:      user.Name,
		UsedName:  user.UsedName,
		Company:   user.Company,
		Birth:     formatDate(user.Birth),
		CreatedAt: user.CreatedAt.Format(http.TimeFormat),
		UpdatedAt: user.UpdatedAt.Format(http.TimeFormat),
	}
}

func formatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.Format("2006-01-02")
	return &formatted
}

func parseUserID(raw string) (uint64, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, errs.NewValidation("invalid user id", map[string]string{
			"id": "id must be an unsigned integer",
		})
	}
	return id, nil
}

func parseIntQuery(c echo.Context, name string) (int, error) {
	value := strings.TrimSpace(c.QueryParam(name))
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errs.NewValidation("invalid query parameter", map[string]string{
			name: name + " must be an integer",
		})
	}
	return parsed, nil
}

func optionalQuery(c echo.Context, name string) *string {
	value := strings.TrimSpace(c.QueryParam(name))
	if value == "" {
		return nil
	}
	return &value
}

func translateValidationError(err error) error {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fields := make(map[string]string, len(validationErrors))
		for _, item := range validationErrors {
			fields[item.Field()] = validationMessage(item)
		}
		return errs.NewValidation("validation failed", fields)
	}

	return errs.NewValidation("validation failed", nil)
}

func validationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "max":
		return "is too long"
	case "datetime":
		return "must use YYYY-MM-DD"
	default:
		return "is invalid"
	}
}
