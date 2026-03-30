package httpapi

// HealthResponse 健康检查响应.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// ErrorDetail 错误详情.
type ErrorDetail struct {
	Code    string            `json:"code" example:"validation_error"`
	Message string            `json:"message" example:"validation failed"`
	Fields  map[string]string `json:"fields,omitempty" swaggertype:"object,string"`
}

// ErrorEnvelope 错误响应包装器.
type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

// Pagination 分页信息.
type Pagination struct {
	Page     int   `json:"page" example:"1"`
	PageSize int   `json:"page_size" example:"20"`
	Total    int64 `json:"total" example:"100"`
}

// UserResponse 用户响应结构.
type UserResponse struct {
	ID        int64   `json:"id" example:"1"`
	UID       string  `json:"uid" example:"user-001"`
	Email     *string `json:"email,omitempty" example:"alice@example.com"`
	Name      string  `json:"name" example:"Alice"`
	UsedName  string  `json:"used_name" example:"Ali"`
	Company   string  `json:"company" example:"ACME"`
	Birth     *string `json:"birth,omitempty" example:"1990-01-01"`
	CreatedAt string  `json:"created_at" example:"Mon, 02 Jan 2006 15:04:05 GMT"`
	UpdatedAt string  `json:"updated_at" example:"Mon, 02 Jan 2006 15:04:05 GMT"`
}

// UserDetailEnvelope 单个用户响应包装器.
type UserDetailEnvelope struct {
	Data UserResponse `json:"data"`
}

// UserListEnvelope 用户列表响应包装器.
type UserListEnvelope struct {
	Data       []UserResponse `json:"data"`
	Pagination Pagination     `json:"pagination"`
}
