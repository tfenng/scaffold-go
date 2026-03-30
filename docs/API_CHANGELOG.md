# API Change Log

## 2026-03-29

本次更新主要影响 `users` 资源的更新语义，前端如有“编辑用户”或“局部保存”能力，请同步调整请求方式。

### 变更概览

- 新增并明确 `PATCH /api/v1/users/{id}` 的真实局部更新语义
- 明确 `PUT /api/v1/users/{id}` 为全量更新语义
- Swagger / OpenAPI 文档已同步更新

### 影响范围

- `PUT /api/v1/users/{id}`
- `PATCH /api/v1/users/{id}`
- Swagger UI
- Swagger JSON / YAML

### 详细变更

#### 1. `PATCH /api/v1/users/{id}` 语义变更

此前：

- `PATCH` 与 `PUT` 复用同一套服务逻辑
- 前端即使只想更新单个字段，也需要按“全量更新”思路提交
- `name` 等字段可能被当作必填处理

现在：

- `PATCH` 只更新请求体中显式出现的字段
- 未出现在请求体中的字段保持原值不变
- 适合前端做“局部保存”“单字段编辑”“按需更新”

示例：

```http
PATCH /api/v1/users/1
Content-Type: application/json

{
  "company": "Example Co"
}
```

效果：

- 仅更新 `company`
- `name`、`email`、`used_name`、`birth` 等其他字段保持不变

#### 2. `PATCH` 现在支持用 `null` 清空可空字段

当前支持显式清空的字段：

- `email`
- `birth`

示例：

```http
PATCH /api/v1/users/1
Content-Type: application/json

{
  "email": null,
  "birth": null
}
```

效果：

- 数据库中的 `email` 被更新为 `NULL`
- 数据库中的 `birth` 被更新为 `NULL`

说明：

- 如果字段未传，不会被清空
- 只有显式传入 `null` 才表示“清空”

#### 3. `PUT /api/v1/users/{id}` 语义澄清

`PUT` 现在明确用于全量更新可变字段。

前端调用 `PUT` 时，应按完整编辑表单提交完整数据，至少包括：

- `name` 必填

建议：

- 完整表单提交使用 `PUT`
- 局部字段修改使用 `PATCH`

### 前端同步建议

#### 如果你们当前“编辑用户”统一走 `PATCH`

请确认请求体是否只发送发生变更的字段：

- 如果是，当前后端语义更匹配，前端代码通常只需要确认字段裁剪逻辑即可
- 如果不是，而是一直发送完整对象，也仍可正常工作，但建议按需精简为局部字段提交

#### 如果你们当前“编辑用户”统一走 `PUT`

请确认：

- 表单提交时始终携带完整可变字段
- `name` 不可缺失

#### 如果前端需要“清空邮箱/生日”

请改为显式发送：

```json
{
  "email": null
}
```

或：

```json
{
  "birth": null
}
```

不要依赖“不传字段”来触发清空，因为“不传”现在表示“保持原值”。

### 兼容性说明

- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `POST /api/v1/users`
- `DELETE /api/v1/users/{id}`

以上接口在本次变更中无协议调整。

### 建议前端核对清单

- 是否区分“全量保存”和“局部保存”
- 是否把 `PATCH` 请求裁剪为仅发送变更字段
- 是否在需要清空 `email` / `birth` 时发送 `null`
- 是否仍将 `PUT` 用于完整对象提交
- 是否已同步阅读最新 Swagger 文档

### 文档入口

- Swagger UI: `http://127.0.0.1:8080/swagger/index.html`
- Swagger JSON: `http://127.0.0.1:8080/swagger/swagger.json`
