# scaffold-api 技术栈概览

## 1. 项目定位

`scaffold-api` 目前是一个以 `users` 资源 CRUD 为核心的 Go REST API 脚手架，目标偏向“简单、显式、工程化”，尽量避免重框架和隐式魔法。

当前主链已经切换为：

`main -> config -> slog -> pgx pool -> sqlc -> service -> chi`

## 2. 核心技术栈总览

| 层次 | 技术/组件 | 在项目中的作用 |
| --- | --- | --- |
| 编程语言 | Go 1.22.x | 项目主语言 |
| 启动方式 | 显式 `main.go` | 负责手动组装 config、logger、db、service、http |
| 配置管理 | 纯环境变量 | 使用 `APP_ENV`、`HTTP_PORT`、`DB_DSN` 等环境变量启动 |
| 日志 | `log/slog` | 统一输出启动日志、请求日志和错误日志 |
| Web 框架 | `chi` | 负责路由、中间件和 HTTP 处理 |
| CORS | `go-chi/cors` | 处理本地前端联调的跨域策略 |
| 参数校验 | `validator/v10` | 对请求体做结构化校验 |
| 数据库 | PostgreSQL | 当前唯一持久化存储 |
| 连接池 | `pgxpool` | 管理 PostgreSQL 连接和生命周期 |
| SQL 代码生成 | `sqlc` | 根据 SQL 文件生成类型安全的查询代码 |
| 数据迁移 | `golang-migrate` | 管理数据库 schema 版本 |
| API 文档 | `swaggo/swag` + `http-swagger` | 生成 Swagger 文档并提供在线 Swagger UI |
| 测试 | `testing` + `testify` | 单元测试和 HTTP 行为测试 |
| 开发命令 | `Makefile` | 提供 `make dev/test/sqlc/migrate/swagger` 等命令 |
| 容器化 | Docker 多阶段构建 | 生成轻量运行镜像 |

## 3. 当前目录职责

当前代码结构重点如下：

- `main.go`
  显式初始化应用依赖并启动 HTTP 服务。
- `internal/config/`
  负责读取环境变量、设置默认值和配置校验。
- `internal/logger/`
  负责初始化 `slog`。
- `internal/db/`
  负责 `pgxpool` 初始化、`sqlc` queries 注入和事务 helper。
- `internal/http/`
  负责 `chi` 路由、中间件、Swagger 路由、错误处理和 Users Handler。
- `internal/service/`
  负责业务规则，如分页、输入清洗、日期解析和数据库错误翻译。
- `db/query/`
  维护 `sqlc` 的 SQL 查询源文件。
- `db/migrations/`
  维护 `golang-migrate` 迁移文件。

整体是一个比较轻量的分层：

`HTTP Handler -> Service -> sqlc/DB`

当前没有保留单独的 repository 层。

## 4. 配置与运行方式

项目已不再依赖 `config.yaml`、`--config`、Cobra 或 Viper。

主要环境变量如下：

- `APP_ENV`
- `HTTP_HOST`
- `HTTP_PORT`
- `DB_DSN`
- `LOG_LEVEL`
- `CORS_ALLOW_ORIGINS`

本地开发常用方式：

- `make dev`
- `./run.sh`

Docker 本地联调方式：

- `./build.sh`
- `./runD.sh`

## 5. HTTP 与接口层

HTTP 层基于 `chi`，当前保留这些能力：

- `RequestID`
- `Recover`
- CORS
- 请求日志
- 统一错误响应
- Swagger UI 与 Swagger JSON

当前暴露的 API 路由仍然保持不变：

- `GET /healthz`
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `PUT /api/v1/users/{id}`
- `PATCH /api/v1/users/{id}`
- `DELETE /api/v1/users/{id}`

## 6. 数据访问与持久化

项目当前的数据层已从 `GORM` 切换为 `pgx + sqlc`。

### 已接入能力

- 使用 `pgxpool` 建立 PostgreSQL 连接
- 使用 `sqlc` 生成 Users CRUD 查询代码
- 使用 `golang-migrate` 管理 `users` 表迁移
- 在 `service` 层做 PostgreSQL 错误翻译
- 提供 `WithTx(ctx, fn)` 事务 helper

### 当前表结构语义

`users` 表包含：

- `id`
- `uid`
- `email`
- `name`
- `used_name`
- `company`
- `birth`
- `created_at`
- `updated_at`

其中 `uid`、`email` 保持唯一性约束。

## 7. 文档与工程化

项目继续保留 Swagger 文档体系：

- `swaggo/swag` 负责从注释生成产物
- `http-swagger` 负责提供 Swagger UI
- 产物位于 `docs/`

常用工程命令通过 `Makefile` 提供：

- `make dev`
- `make build`
- `make test`
- `make sqlc`
- `make migrate-up`
- `make migrate-down`
- `make swagger`
- `make swagger-check`

## 8. 当前边界与剩余事项

当前主栈已经迁移完成，但仍有一些收尾任务可以继续推进：

- 将 `TODO.md` 中已完成项勾选
- 继续完善集成测试，尤其是带 PostgreSQL 的真实迁移与 CRUD 验证
- 清理历史说明文档中对旧栈的描述，例如 `PRD.md`

## 9. 一句话总结

这是一个基于 `chi + slog + pgxpool + sqlc + golang-migrate + Swagger + Makefile` 的 Go REST API 脚手架，强调显式初始化、轻量分层、类型安全 SQL 和可维护的工程流程。
