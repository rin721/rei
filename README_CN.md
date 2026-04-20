# rei

`rei` 是一个面向中大型 Web 服务的模块化 Go 后端脚手架。

## 当前结构

- `cmd/` 暴露统一 CLI：`run` 与 `db {generate|migrate|status|rollback}`
- `internal/config` 负责类型化配置、环境变量覆盖与热重载
- `internal/app` 负责运行时与数据库管理入口的装配
- `internal/models`、`internal/repository`、`internal/service`、`internal/handler` 组成当前业务主链路
- `scripts/migrations/` 是唯一的数据库结构历史来源

## 模块路径

```text
github.com/rin721/rei
```

## 快速开始

```bash
go list ./...
go test ./...
go vet ./...
go run ./cmd run --dry-run
go run ./cmd db migrate --dry-run
go run ./cmd run
```

默认示例配置使用本地 SQLite，可以在没有外部依赖的情况下完成本地验证。

## 迁移工作流

- 使用 `go run ./cmd db generate --desc <name>` 生成版本化 SQL
- 在 `scripts/migrations/` 中审阅迁移脚本
- 使用 `go run ./cmd db migrate` 应用待执行迁移
- 使用 `go run ./cmd db status` 查看当前迁移状态

## HTTP 路由

- `GET /health`
- `GET /api/v1/samples`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/change-password`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/users/me`
- `PUT /api/v1/users/me`
- `GET /api/v1/rbac/check`
- `POST /api/v1/rbac/roles/assign`
- `POST /api/v1/rbac/roles/revoke`
- `GET /api/v1/rbac/users/:user_id/roles`
- `GET /api/v1/rbac/roles/:role/users`
- `POST /api/v1/rbac/policies`
- `DELETE /api/v1/rbac/policies`
- `GET /api/v1/rbac/policies`

## 说明

- YAML 键统一使用 `snake_case`
- 环境变量统一使用 `UPPER_SNAKE_CASE`
- API 响应统一使用 `code`、`message`、`data`、`traceId`、`serverTime`
- 密码哈希不会出现在 JSON 响应中
