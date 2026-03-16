# scaffold-api

基于 `Cobra + Viper + Fx + Echo + GORM + Zerolog` 的最小 Go 后端脚手架，目前只实现 `users` 数据模型的 RESTful CRUD 管理接口，并集成在线 Swagger 文档。

## 功能概览

- `users` CRUD 接口
- PostgreSQL + GORM 持久化
- `Cobra` 主应用入口
- `Viper` 配置加载
- 在线 Swagger UI
- 机器可读 Swagger JSON

## 环境要求

- Go `1.22.2+`
- PostgreSQL

## 主应用入口

程序入口是：

- `main.go`
- `scaffold-api serve`

查看命令帮助：

```bash
go run . --help
go run . serve --help
```

启动 HTTP 服务：

```bash
go run . serve --config configs/config.yaml
```

也可以直接通过 flag 启动：

```bash
go run . serve \
  --db-dsn 'postgres://postgres:postgres@127.0.0.1:5432/scaffold_api?sslmode=disable' \
  --http-port 8080
```

编译后二进制启动方式：

```bash
go build -o scaffold-api .
./scaffold-api serve --config configs/config.yaml
```

## 配置说明

项目使用 `Viper` 加载配置，优先级如下：

`命令行 Flag > 环境变量 > 配置文件 > 默认值`

默认会尝试读取：

- 当前目录下的 `config.yaml`
- `configs/config.yaml`

也可以通过 `--config` 指定路径。

配置样例见：

- `configs/config.yaml.example`

当前支持的主要配置项：

| Key | 说明 | 默认值 |
| --- | --- | --- |
| `app_name` | 应用名 | `scaffold-api` |
| `environment` | 运行环境 | `dev` |
| `http_host` | HTTP 监听地址 | `0.0.0.0` |
| `http_port` | HTTP 监听端口 | `8080` |
| `http_read_timeout_seconds` | HTTP 读超时 | `15` |
| `http_write_timeout_seconds` | HTTP 写超时 | `15` |
| `http_shutdown_timeout_seconds` | HTTP 关闭超时 | `10` |
| `db_dsn` | PostgreSQL DSN | 必填 |
| `db_auto_migrate` | 启动时自动迁移 | `true` |
| `log_level` | 日志级别 | `info` |
| `log_pretty` | 控制台友好日志 | `true` |

环境变量前缀为 `APP_`，例如：

```bash
export APP_DB_DSN='postgres://postgres:postgres@127.0.0.1:5432/scaffold_api?sslmode=disable'
export APP_HTTP_PORT=8080
go run . serve
```

## 在线 API 文档

启动服务后，可访问：

- Swagger UI: `http://127.0.0.1:8080/swagger/index.html`
- Swagger JSON: `http://127.0.0.1:8080/swagger/swagger.json`

说明：

- Swagger UI 适合前端、测试或联调直接查看和在线调试
- Swagger JSON 适合前端工具链、API 平台或网关做机器消费

## 生成 Swagger 文档

项目已经提交了生成产物：

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

如果修改了接口注释或请求/响应结构体，需要重新生成：

```bash
go generate ./...
```

`go generate` 依赖本地安装 `swag`。首次使用可执行：

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
```

也可以直接手动生成：

```bash
swag init --parseInternal -g main.go -o docs
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

## 开发与验证

运行测试：

```bash
go test ./...
```

构建应用：

```bash
go build ./...
```
