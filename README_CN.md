# go-scaffold2

`go-scaffold2` 是一个按阶段推进的模块化 Go 后端脚手架，面向中大型 Web 服务场景。

当前仓库已经完成 Phase 0 到 Phase 7：

- 根目录骨架、工程文件与安全配置示例
- 共享常量、错误体系、统一响应 envelope 与基础契约
- 可复用的 `pkg/*` 基础设施层
- `internal/config` 配置加载、环境变量覆盖与热重载
- `internal/app` 的 `run` / `initdb` 容器装配
- `internal/middleware` 与 `internal/router` 的完整 Web 链路
- `internal/models`、`internal/repository`、`internal/service`、`internal/handler`
- 已经打通的 auth、user、rbac、sample 业务模块
- 可执行的 `initdb` SQL 生成、lock file 保护与同步文档
- 已完成最终质量收口，并通过 `go test ./...` 与 `go vet ./...`

## 当前阶段

项目目前已经进入 Phase 7 的可验收状态。规划内的运行时、业务主链路和 `initdb` 工作流都已经接通并验证完成。

## 模块路径

```text
github.com/rei0721/go-scaffold2
```

## 快速验证

```bash
go list ./...
go test ./...
go vet ./...
go run ./cmd/server run --dry-run
go run ./cmd/server initdb --dry-run
go run ./cmd/server initdb
go run ./cmd/server
```

默认示例配置已经切到安全的本地 SQLite，开箱即可运行，不依赖外部数据库。

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

## InitDB 输出

运行 `initdb` 后会写入：

- `scripts/initdb/initdb.<driver>.sql`
- 非 dry-run 成功后额外写入 `scripts/initdb/.initdb.lock`

## 当前约定

- YAML 键统一使用 `snake_case`
- 环境变量统一使用 `UPPER_SNAKE_CASE`
- API 响应 envelope 固定为 `code`、`message`、`data`、`traceId`、`serverTime`
- 示例配置只包含占位符和安全默认值
- 密码哈希不会出现在任何 JSON 响应中
