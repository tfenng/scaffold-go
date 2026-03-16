package server

// HealthResponse describes the health check payload.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// ErrorDetail describes a normalized API error body.
type ErrorDetail struct {
	Code    string            `json:"code" example:"validation_error"`
	Message string            `json:"message" example:"validation failed"`
	Fields  map[string]string `json:"fields,omitempty" swaggertype:"object,string"`
}

// ErrorEnvelope wraps an API error response.
type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

// Pagination describes paging metadata for list responses.
type Pagination struct {
	Page     int   `json:"page" example:"1"`
	PageSize int   `json:"page_size" example:"20"`
	Total    int64 `json:"total" example:"100"`
}

// UserResponse describes a user returned by the API.
type UserResponse struct {
	ID        uint64  `json:"id" example:"1"`
	UID       string  `json:"uid" example:"user-001"`
	Email     *string `json:"email,omitempty" example:"alice@example.com"`
	Name      string  `json:"name" example:"Alice"`
	UsedName  string  `json:"used_name" example:"Ali"`
	Company   string  `json:"company" example:"ACME"`
	Birth     *string `json:"birth,omitempty" example:"1990-01-01"`
	CreatedAt string  `json:"created_at" example:"Mon, 02 Jan 2006 15:04:05 GMT"`
	UpdatedAt string  `json:"updated_at" example:"Mon, 02 Jan 2006 15:04:05 GMT"`
}

// UserDetailEnvelope wraps a single user response.
type UserDetailEnvelope struct {
	Data UserResponse `json:"data"`
}

// UserListEnvelope wraps a paginated users response.
type UserListEnvelope struct {
	Data       []UserResponse `json:"data"`
	Pagination Pagination     `json:"pagination"`
}
