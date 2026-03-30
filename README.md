# scaffold-api

基于 `chi + slog + pgx + sqlc + Swagger` 的最小 Go 后端脚手架，当前实现 `users` 资源的 RESTful CRUD，并保留在线 Swagger 文档。

## 功能概览

- `users` CRUD 接口
- PostgreSQL 持久化
- `sqlc` 生成查询代码
- `golang-migrate` 管理数据库迁移
- 在线 Swagger UI
- 机器可读 Swagger JSON
- `Makefile` 驱动常用开发命令

## 环境要求

- Go `1.22.2+`
- PostgreSQL
- 可选工具：
  - `sqlc`
  - `migrate`
  - `swag`

## 快速开始

复制环境变量模板：

```bash
cp .env.sample .env.dev
```

根据本地 PostgreSQL 实际情况修改 `.env.dev` 中的 `DB_DSN`，然后启动服务：

```bash
make dev
```

也可以直接使用脚本：

```bash
./run.sh
```

## 配置说明

项目改为纯环境变量配置，不再依赖 `config.yaml` 和 `--config`。

当前主要配置项：

| Key | 说明 | 默认值 |
| --- | --- | --- |
| `APP_ENV` | 运行环境 | `dev` |
| `HTTP_HOST` | HTTP 监听地址 | `0.0.0.0` |
| `HTTP_PORT` | HTTP 监听端口 | `8080` |
| `DB_DSN` | PostgreSQL DSN | 必填 |
| `LOG_LEVEL` | 日志级别 | `info` |
| `CORS_ALLOW_ORIGINS` | 允许的前端来源，逗号分隔 | `http://localhost:3000,http://127.0.0.1:3000` |

示例：

```bash
export APP_ENV=dev
export HTTP_PORT=8080
export DB_DSN='postgres://postgres:postgres@127.0.0.1:5432/scaffold_api?sslmode=disable'
go run .
```

## 运行与开发命令

常用命令通过 `Makefile` 提供：

```bash
make dev
make build
make test
make sqlc
make migrate-up
make migrate-down
make swagger
make swagger-check
```

说明：

- `make dev` 会默认加载 `.env.dev`
- `make sqlc` 依赖本地已安装 `sqlc`
- `make migrate-up` / `make migrate-down` 依赖当前 shell 已设置 `DB_DSN`
- `make swagger` 会执行 `go generate ./...`

## 数据库迁移

迁移文件位于：

- `db/migrations/`

执行迁移前，请先导出 `DB_DSN`：

```bash
export DB_DSN='postgres://postgres:postgres@127.0.0.1:5432/scaffold_api?sslmode=disable'
make migrate-up
```

回滚一版迁移：

```bash
make migrate-down
```

## SQL 与数据访问

`sqlc` 配置文件位于：

- `sqlc.yaml`

查询源文件位于：

- `db/query/users.sql`

生成代码位于：

- `internal/db/query/`

重新生成查询代码：

```bash
make sqlc
```

## 架构与请求流

启动入口：

- `main.go` 负责加载配置、初始化日志、连接 PostgreSQL，并组装 `store -> service -> http handler`

请求路径：

1. `internal/http/router.go` 注册路由、中间件、Swagger 和 `users` API
2. `internal/http/user_handler.go` 负责解析请求、校验输入、写出响应
3. `internal/service/user_service.go` 负责业务规则、分页默认值、字段归一化、错误翻译
4. `internal/db/store.go` 作为存储适配层，调用 `sqlc` 生成的方法
5. `db/query/users.sql` 定义最终执行的 SQL

错误处理：

- Handler 返回 `error`
- `internal/http/errors.go` 统一把应用错误转换为 HTTP 状态码和 JSON 错误响应
- `internal/http/middleware.go` 记录 request id、状态码和耗时

## 在线 API 文档

启动服务后，可访问：

- Swagger UI: `http://127.0.0.1:8080/swagger/index.html`
- Swagger JSON: `http://127.0.0.1:8080/swagger/swagger.json`

Swagger 生成产物已提交到仓库：

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`
- `docs/API_CHANGELOG.md`

如果修改了接口注释或请求/响应结构体，需要重新生成：

```bash
make swagger
```

校验 Swagger 产物是否已同步：

```bash
make swagger-check
```

## 接口清单

当前已实现接口：

- `GET /healthz`
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `PUT /api/v1/users/{id}`
- `PATCH /api/v1/users/{id}`
- `DELETE /api/v1/users/{id}`

更新语义：

- `PUT /api/v1/users/{id}`: 全量更新可变字段，`name` 必填
- `PATCH /api/v1/users/{id}`: 部分更新，只修改请求体中出现的字段
- `PATCH` 中显式传入 `null` 可清空可空字段，例如 `email` 和 `birth`

## 接口示例

以 `PATCH /api/v1/users/{id}` 为例：

请求：

```http
PATCH /api/v1/users/1
Content-Type: application/json

{
  "company": "Example Co",
  "email": null
}
```

含义：

- 仅更新 `company`
- 将 `email` 清空为 `NULL`
- 其他字段保持原值不变

成功响应：

```json
{
  "data": {
    "id": 1,
    "uid": "user-001",
    "email": null,
    "name": "Alice",
    "used_name": "Ali",
    "company": "Example Co",
    "birth": "1990-01-01",
    "created_at": "Mon, 02 Jan 2006 15:04:05 GMT",
    "updated_at": "Mon, 02 Jan 2006 15:04:05 GMT"
  }
}
```

执行链路：

1. `router.go` 将请求分发到 `UserHandler.Patch`
2. `user_handler.go` 解码 JSON 并调用 `UserService.Patch`
3. `user_service.go` 先读取当前用户，再将补丁字段合并到现有数据
4. `store.go` 调用 `sqlc` 生成的 `UpdateUser`
5. `users.sql` 执行 `UPDATE users ... RETURNING ...`

## 测试与构建

运行测试：

```bash
make test
```

构建应用：

```bash
make build
```

构建 Docker 镜像：

```bash
./build.sh
```

## 开发期临时启动 Docker 容器

默认流程：

```bash
./build.sh
./run.sh
```

说明：

- `run.sh` 默认读取 `.env.dev`
- 如果 `DB_DSN` 中使用了 `localhost` 或 `127.0.0.1`，脚本会自动替换为 `host.docker.internal`
- Linux 下会自动增加 `host.docker.internal:host-gateway` 映射，方便容器访问宿主机 PostgreSQL
