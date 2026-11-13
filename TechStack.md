# scaffold-api 技术栈概览

## 1. 项目定位

`scaffold-api` 是一个以 `users` 资源 CRUD 为核心的 Go 后端脚手架，强调“显式、轻量、单二进制可部署”。前端静态产物内嵌进后端，整个应用以一个进程启动。

当前主链：

`main -> config -> slog -> sql.DB(turso) -> migrations -> store/sqlc -> service -> chi(+webui)`

## 2. 核心技术栈

| 层次 | 技术/组件 | 作用 |
| --- | --- | --- |
| 语言 | Go 1.24+ | 项目主语言 |
| 启动 | 显式 `main.go` | 手动装配 config、logger、db、service、http |
| 配置 | 纯环境变量 | `APP_ENV`、`HTTP_PORT`、`DB_PATH` 等 |
| 日志 | `log/slog` | 结构化启动日志、请求日志、错误日志 |
| Web 框架 | `chi` | 路由、中间件、静态资源兜底 |
| CORS | `go-chi/cors` | 前端联调跨域 |
| 校验 | `validator/v10` | 请求体结构校验 |
| 数据库 | Turso（SQLite 兼容，嵌入式） | 进程内持久化，无外部服务 |
| 驱动 | `turso.tech/database/tursogo` | 实现 `database/sql`，免 CGO |
| 连接管理 | `database/sql` | 单连接池（SQLite 单写者） |
| SQL 生成 | `sqlc`（engine: sqlite） | 生成类型安全查询 |
| 迁移 | 内置 migrator + `go:embed` | 启动自动迁移，也可 `cmd/migrate` |
| API 文档 | `swaggo/swag` + `http-swagger` | Swagger 2.0 文档与 UI |
| 前端 | Next.js 静态导出 + `go:embed` | 单二进制托管 Web UI |
| 测试 | `testing` + `testify` | 单元测试与内存库集成测试 |
| 构建 | `Makefile` | `dev/web/release/test/sqlc/migrate/swagger` |

## 3. 目录职责

- `main.go`：初始化依赖并启动 HTTP 服务，启动时自动迁移。
- `cmd/migrate/`：手动执行迁移（up/down）。
- `internal/config/`：读取环境变量、默认值与校验。
- `internal/logger/`：初始化 `slog`。
- `internal/db/`：打开 Turso、迁移器、`Store`、事务 helper。
- `internal/db/query/`：sqlc 生成代码（勿手改）。
- `internal/http/`：`chi` 路由、中间件、Swagger、Users Handler、静态资源兜底。
- `internal/service/`：业务规则、分页、字段归一化、错误翻译。
- `internal/webui/`：`go:embed` 前端构建产物并提供静态处理器。
- `db/query/`：sqlc SQL 源文件。
- `db/migrations/`：迁移 SQL 与 embed 声明。

分层：`HTTP Handler -> Service -> Store/sqlc -> Turso`。

## 4. 配置

主要环境变量：

- `APP_ENV`
- `HTTP_HOST`
- `HTTP_PORT`
- `DB_PATH`（文件路径或 `:memory:`，默认 `./data/scaffold.db`）
- `LOG_LEVEL`
- `CORS_ALLOW_ORIGINS`

## 5. HTTP 与静态资源

中间件与能力：

- `RequestID`、`Recover`、CORS、请求日志、统一错误响应
- Swagger UI 与 Swagger JSON
- 前端静态资源兜底（`internal/webui`）

路由：

- `GET /healthz`
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `PUT /api/v1/users/{id}`
- `PATCH /api/v1/users/{id}`
- `DELETE /api/v1/users/{id}`
- `/*`：前端静态资源

## 6. 数据访问与持久化

- 使用 `turso.tech/database/tursogo` + `database/sql` 打开进程内数据库。
- `sqlc` 以 `engine: sqlite` / `sql_package: database/sql` 生成查询。
- sqlc 类型：可空列为 `*string`，整数为 `int64`，时间字段为 RFC3339 文本。
- 迁移文件通过 `go:embed` 嵌入，启动时自动应用；`schema_migrations` 记录版本。
- 错误翻译：`sql.ErrNoRows -> 404`，`turso.ErrTursoConstraint -> 409`。
- 事务 helper：`Store.WithTx(ctx, fn)`。

`users` 表包含：`id`、`uid`、`email`、`name`、`used_name`、`company`、`birth`、`created_at`、`updated_at`；`uid`、`email` 唯一。

## 7. 单二进制与前端

- 前端 `scaffold-web` 使用 Next.js `output: "export"`。
- `make web` 将 `scaffold-web/out` 复制到 `internal/webui/dist`。
- `internal/webui` 通过 `go:embed all:dist` 打包，并在 `chi` 中以 `/*` 兜底。
- `make release` 产出 `bin/scaffold-api`。
- 未构建前端时，根路径返回 503 提示。

## 8. 工程命令

- `make dev` / `make dev-web`
- `make build`
- `make web`
- `make release`
- `make test`
- `make sqlc`
- `make migrate-up` / `make migrate-down`
- `make swagger` / `make swagger-check`

## 9. 一句话总结

这是一个 `chi + slog + database/sql + Turso(SQLite) + sqlc`、并把前端静态产物内嵌进同一二进制的 Go 全栈脚手架，强调显式装配、类型安全 SQL、零外部数据库依赖与单文件部署。
