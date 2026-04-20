# Clean Architecture Constitution

## 适用范围

本规则适用于整个仓库内的所有业务模块、基础设施适配器、HTTP 入口与未来重构产物。

## 总则

1. 项目必须坚持严格的 Clean Architecture 与 DDD。
2. 业务代码必须面向 Domain 与 Ports 编程。
3. 基础设施实现只能存在于边缘层，不得向 Usecase 和 Domain 反向渗透。
4. 任何新模块都必须服从本分层协议，旧代码重构时也必须向本协议收敛。

## Domain / Entity

1. Domain 实体必须是纯粹的 Go 结构体。
2. Domain 实体禁止携带任何 `gorm` 标签。
3. Domain 实体禁止携带任何 `json` 标签。
4. Domain 实体禁止依赖任何外部库。
5. Domain 层禁止感知 HTTP、数据库、缓存、消息队列、配置系统、权限引擎等基础设施细节。
6. Domain 只表达业务语义、业务状态、领域行为与领域约束。

## Usecase / Service

1. Usecase 层只承载纯业务逻辑。
2. Usecase 层必须面向 Domain 实体编程。
3. Usecase 层严禁直接引用 `*gorm.DB`。
4. Usecase 层严禁直接引用 `pkg/jwt`、`pkg/rbac`、`pkg/cache`、`pkg/database` 等基础设施类型。
5. Usecase 层必须通过接口（Ports）与外部系统通信。
6. Usecase 的输入与输出应是 Domain 实体、领域值对象或专门定义的应用层结果对象。
7. Usecase 严禁直接返回 ORM Model。
8. 事务能力必须抽象为 Port，事务实现细节不得泄漏到业务逻辑中。

## Repository / Adapter

1. Repository / Adapter 层负责连接业务世界与基础设施世界。
2. 该层负责 Domain 实体与底层模型之间的双向转换。
3. GORM 模型必须隐藏在 Repository / Adapter 内部。
4. Redis、JWT、RBAC、第三方 SDK、文件存储等具体实现，也必须隐藏在 Adapter 内部。
5. Adapter 可以依赖 GORM、Redis、Casbin、第三方 SDK，但不得把这些类型暴露给 Usecase。
6. Repository 接口应定义在 Usecase 所依赖的 Port 层，具体实现位于 Adapter 层。

## Delivery / Handler

1. Delivery 层负责处理 HTTP 请求与响应。
2. Delivery 层负责 DTO 解析、参数校验、错误翻译与响应封装。
3. Delivery 层负责 DTO 与 Domain 实体之间的转换。
4. Delivery 层只能调用 Usecase，不得直接调用 GORM、Redis、RBAC、JWT Manager 等基础设施实现。
5. Delivery 层不得承载业务规则。
6. Handler 中不得编写事务逻辑、拼接 SQL 或执行权限持久化操作。

## 依赖方向

必须遵守以下依赖方向：

`Delivery -> Usecase -> Domain`

`Adapter -> Domain`

`Top-level App / Composition Root -> Delivery + Usecase + Adapter`

严格禁止以下情况：

1. Domain 依赖 Usecase、Delivery 或 Adapter。
2. Usecase 依赖 Gin、GORM、Redis、JWT、Casbin 或具体 HTTP Client 实现。
3. Delivery 依赖 Repository 的具体实现细节。
4. GORM Model 穿透到 Usecase 或 Delivery。

## 三类对象隔离

1. Domain Entity 是业务真相。
2. DTO 是 Delivery 边界对象。
3. ORM Model 是持久化边界对象。
4. 三者不得混用。
5. 禁止在同一结构体上同时承担 Domain、DTO、ORM Model 三种职责。

## 审查硬性红线

出现以下任一情况，必须视为架构违规：

1. Usecase 直接引用 `*gorm.DB`。
2. Usecase 直接引用 `pkg/jwt`、`pkg/rbac` 等基础设施类型。
3. Domain 结构体携带 `gorm` 或 `json` 标签。
4. Handler 直接操作数据库、缓存或其他基础设施实现。
5. GORM Model 暴露到 Usecase 或 Handler。
