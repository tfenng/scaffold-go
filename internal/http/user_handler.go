package httpapi

import (
	"encoding/json"
	"errors"
	"math"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"scaffold-api/internal/db/query"
	"scaffold-api/internal/errs"
	"scaffold-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// UserHandler handles user CRUD requests.
type UserHandler struct {
	service  *service.UserService
	validate *validator.Validate
}

// NewUserHandler creates a user handler.
func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{
		service:  service,
		validate: validator.New(),
	}
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
func (h *UserHandler) Create(w stdhttp.ResponseWriter, r *stdhttp.Request) error {
	var req service.CreateUserInput
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if err := h.validate.Struct(&req); err != nil {
		return translateValidationError(err)
	}

	user, err := h.service.Create(r.Context(), req)
	if err != nil {
		return err
	}

	writeJSON(w, stdhttp.StatusCreated, UserDetailEnvelope{Data: newUserResponse(*user)})
	return nil
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
func (h *UserHandler) GetByID(w stdhttp.ResponseWriter, r *stdhttp.Request) error {
	id, err := parseUserID(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		return err
	}

	writeJSON(w, stdhttp.StatusOK, UserDetailEnvelope{Data: newUserResponse(*user)})
	return nil
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
func (h *UserHandler) List(w stdhttp.ResponseWriter, r *stdhttp.Request) error {
	page, err := parseIntQuery(r, "page")
	if err != nil {
		return err
	}
	pageSize, err := parseIntQuery(r, "page_size")
	if err != nil {
		return err
	}

	result, err := h.service.List(r.Context(), service.ListUsersInput{
		Email:    optionalQuery(r, "email"),
		NameLike: optionalQuery(r, "name_like"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return err
	}

	items := make([]UserResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, newUserResponse(item))
	}

	writeJSON(w, stdhttp.StatusOK, UserListEnvelope{
		Data: items,
		Pagination: Pagination{
			Page:     result.Page,
			PageSize: result.PageSize,
			Total:    result.Total,
		},
	})
	return nil
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
func (h *UserHandler) Update(w stdhttp.ResponseWriter, r *stdhttp.Request) error {
	id, err := parseUserID(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}

	var req service.UpdateUserInput
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if err := h.validate.Struct(&req); err != nil {
		return translateValidationError(err)
	}

	user, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		return err
	}

	writeJSON(w, stdhttp.StatusOK, UserDetailEnvelope{Data: newUserResponse(*user)})
	return nil
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
func (h *UserHandler) Delete(w stdhttp.ResponseWriter, r *stdhttp.Request) error {
	id, err := parseUserID(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		return err
	}

	w.WriteHeader(stdhttp.StatusNoContent)
	return nil
}

func decodeJSON(r *stdhttp.Request, dest any) error {
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return errs.NewInvalidArgument("invalid request body", map[string]string{
			"body": "request body is invalid",
		})
	}

	return nil
}

func newUserResponse(user query.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		UID:       user.Uid,
		Email:     user.Email,
		Name:      user.Name,
		UsedName:  user.UsedName,
		Company:   user.Company,
		Birth:     formatDate(user.Birth),
		CreatedAt: formatTimestamp(user.CreatedAt.Time),
		UpdatedAt: formatTimestamp(user.UpdatedAt.Time),
	}
}

func formatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.Format("2006-01-02")
	return &formatted
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(stdhttp.TimeFormat)
}

func parseUserID(raw string) (int64, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 || id > math.MaxInt64 {
		return 0, errs.NewInvalidArgument("invalid user id", map[string]string{
			"id": "id must be an unsigned integer",
		})
	}
	return int64(id), nil
}

func parseIntQuery(r *stdhttp.Request, name string) (int, error) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errs.NewInvalidArgument("invalid query parameter", map[string]string{
			name: name + " must be an integer",
		})
	}
	return parsed, nil
}

func optionalQuery(r *stdhttp.Request, name string) *string {
	value := strings.TrimSpace(r.URL.Query().Get(name))
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
		return errs.NewInvalidArgument("validation failed", fields)
	}

	return errs.NewInvalidArgument("validation failed", nil)
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
