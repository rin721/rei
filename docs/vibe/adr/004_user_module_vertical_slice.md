# ADR 004: Use User as the First Clean Architecture Vertical Slice

日期: 2026-04-20

状态: Accepted

## 背景

在切除旧版 `initdb` 和运行时 `AutoMigrate` 之后，仓库的下一阶段目标是把业务代码真正收敛到 Clean Architecture。

现状里，`service` 层仍然普遍直接依赖 `internal/models`、仓储实现细节以及 Delivery DTO。为了降低一次性重构的风险，我们决定先选一个复杂度适中、外部影响面可控的业务模块做样板切片，再把同样模式复制到更重的 `auth` 和 `rbac`。

## 决策

1. 选择 `user` 模块作为第一块架构样板切片。
2. 新增纯领域实体 `internal/domain/user.User`，不带 `gorm` 或 `json` 标签。
3. 将 `user usecase` 的输入输出改为应用层对象，而不是直接使用 HTTP DTO。
4. 由 `handler/user_handler.go` 负责：
   - 将 `types/user` 中的 HTTP DTO 转为 usecase command/query。
   - 将 usecase result 转回 HTTP response DTO。
5. 由 `repository` 层新增 adapter，负责 `Domain User <-> GORM User Model` 的转换。
6. 对外保持兼容：
   - 不改路由。
   - 不改 JSON 字段。
   - 不改状态码与响应包裹格式。

## 结果

1. `internal/service/user` 已不再直接依赖 `internal/models`。
2. `internal/service/user` 已不再直接依赖 `internal/repository`。
3. `internal/service/user` 已不再直接依赖 `types/user` 的 HTTP DTO。
4. `user` 模块形成了后续 `auth` 和 `rbac` 可复用的重构模板。

## 守卫规则

- 新增业务逻辑时，不得重新让 `user usecase` 直接返回 HTTP DTO。
- 新增持久化逻辑时，不得重新让 `internal/models.User` 穿透到 `user handler` 或 `user usecase`。
- 后续模块重构优先复用这次的切片模式，而不是再引入新的混合层次。
