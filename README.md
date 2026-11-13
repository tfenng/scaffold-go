# scaffold-api

基于 `chi + log/slog + database/sql + Turso（SQLite 兼容）+ sqlc` 的 Go 后端脚手架。前端（Next.js 静态导出）构建产物通过 `go:embed` 打入后端，最终以**单个二进制**启动，无需独立数据库或静态文件服务器。

## 功能概览

- `users` 资源 RESTful CRUD
- 进程内 Turso 数据库（SQLite 文件格式兼容），零外部依赖
- `sqlc` 生成类型安全查询（`engine: sqlite`，`database/sql`）
- 内置迁移器，启动时自动建表/升级
- 前端静态产物内嵌，单二进制同时托管 API 与 Web UI
- 在线 Swagger UI 与 Swagger JSON
- `Makefile` 驱动常用开发命令

## 环境要求

- Go `1.24+`
- 构建前端时才需要：Node.js `18.17+` / `20+`，`pnpm 8+`

数据库是进程内的，不再需要 PostgreSQL 或 Docker。

## 快速开始

```bash
cp env.sample .env.dev
# 按需修改 DB_PATH，默认 ./data/scaffold.db
make dev
```

服务启动时会自动执行迁移并创建 `DB_PATH` 指向的数据库文件。

### 单二进制发布

```bash
make web       # 构建前端，并把 out/ 复制到 internal/webui/dist
make release   # 产出 bin/scaffold-api（API + 前端 + 数据库驱动）
./bin/scaffold-api
```

> `DB_PATH=:memory:` 可启动一个临时内存库，适合演示或测试。

## 配置说明

项目使用纯环境变量配置，不依赖 `config.yaml` 或 `--config`。

| Key | 说明 | 默认值 |
| --- | --- | --- |
| `APP_ENV` | 运行环境 | `dev` |
| `HTTP_HOST` | HTTP 监听地址 | `0.0.0.0` |
| `HTTP_PORT` | HTTP 监听端口 | `8080` |
| `DB_PATH` | Turso/SQLite 数据库文件路径，或 `:memory:` | `./data/scaffold.db` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `CORS_ALLOW_ORIGINS` | 允许的前端来源，逗号分隔 | `http://localhost:3000,http://127.0.0.1:3000` |

示例：

```bash
export APP_ENV=dev
export HTTP_PORT=8080
export DB_PATH=./data/scaffold.db
go run .
```

## 运行与开发命令

```bash
make dev          # 加载 .env.dev，本地运行 API
make dev-web      # 运行 Next.js 开发服务器（前端热更新）
make build        # 编译检查：go build ./...
make web          # 构建前端并复制到 internal/webui/dist
make release      # 构建单二进制
make test         # go test ./...
make sqlc         # 重新生成 sqlc 查询代码
make migrate-up   # 应用迁移
make migrate-down # 回滚最近一次迁移
make swagger      # 重新生成 Swagger 文档
make swagger-check
```

### Windows（PowerShell）

没有安装 `make` 时，使用等价脚本：

```powershell
.\dev.ps1        # 等价于 make dev（读取 .env.dev）
.\build.ps1      # 等价于 make web + make release
go test ./...    # 等价于 make test
go run ./cmd/migrate up   # 等价于 make migrate-up
```

首次构建前端前需要安装 Node.js 与 pnpm：

```powershell
winget install OpenJS.NodeJS.LTS
npm install -g pnpm
```

> 新版 Node 已不再随附 `corepack`，所以用 `npm install -g pnpm`；不装 pnpm 时 `build.ps1` 会自动改用 npm。

如果 PowerShell 阻止脚本执行，可这样运行：

```powershell
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

## 数据库与迁移

- 数据库为嵌入式 Turso（SQLite 兼容），通过 `turso.tech/database/tursogo` 访问。
- 迁移文件位于 `db/migrations/`，命名 `000001_x.up.sql` / `000001_x.down.sql`。
- 迁移文件通过 `db/migrations/embed.go` 嵌入二进制，启动时自动执行 `MigrateUp`。
- 也提供 `cmd/migrate`：

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
```

注意事项：

- SQLite/Turso 是单写者模型，连接池被限制为 1 条连接，避免 `database is busy`。
- 多副本部署时每个实例会拥有各自的数据库文件；需要共享数据时应共享存储或改为单实例/外部数据库方案。
- 时间字段以 RFC3339（UTC）文本存储与返回。

## 单二进制与前端

- `scaffold-web` 是 Next.js 项目，已启用 `output: "export"`，产物在 `out/`。
- `make web` 会把 `scaffold-web/out` 同步到 `internal/webui/dist`。
- `internal/webui` 使用 `go:embed all:dist` 打包静态资源，并在 `chi` 路由中以 `/*` 兜底托管。
- 路由优先级：`/api/v1/*`、`/swagger/*`、`/healthz`，其余请求回落到前端。
- 若未执行 `make web`，根路径会返回 503 并提示先构建前端。

## 架构与请求流

启动入口：

- `main.go` 加载配置、初始化日志、打开数据库、执行迁移，并组装 `store -> service -> http handler`

请求路径：

1. `internal/http/router.go` 注册路由、中间件、Swagger、API 与前端静态资源
2. `internal/http/user_handler.go` 解析请求、校验输入、写出响应
3. `internal/service/user_service.go` 业务规则、分页、字段归一化、错误翻译
4. `internal/db/store.go` 调用 `sqlc` 生成的查询
5. `db/query/users.sql` 定义最终执行的 SQL

## 接口清单

- `GET /healthz`
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `PUT /api/v1/users/{id}`
- `PATCH /api/v1/users/{id}`
- `DELETE /api/v1/users/{id}`

更新语义：

- `PUT`：全量更新可变字段，`name` 必填
- `PATCH`：部分更新，只修改请求体中出现的字段
- `PATCH` 中显式传入 `null` 可清空可空字段，例如 `email` 和 `birth`

## 在线 API 文档

启动服务后可访问：

- Swagger UI: `http://127.0.0.1:8080/swagger/index.html`
- Swagger JSON: `http://127.0.0.1:8080/swagger/swagger.json`

修改接口注释或请求/响应结构体后重新生成：

```bash
make swagger
```

## 测试与构建

```bash
make test
make build
```

`internal/db` 中包含基于内存数据库的集成测试，覆盖迁移、CRUD 与唯一约束错误映射。
