# ADR 005: Refactor Modules Through Backward-Compatible Vertical Slices

日期: 2026-04-20

状态: Accepted

## 背景

`user` 模块的第一轮重构证明了一件事：我们不需要等待“大爆炸式重写”才能推进 Clean Architecture。只要控制好边界，就可以在保持 HTTP API、数据库结构和运行行为稳定的前提下，逐模块地把内部实现迁移到 `Domain + Ports + Usecase + Adapter`。

这次实践后，我们形成了一条更通用的工程共识：后续 `auth`、`rbac` 等模块也应当采用同样的纵切迁移策略，而不是在全仓库同时展开高风险改造。

## 决策

1. 后续架构重构采用“纵切迁移”策略，按模块逐步推进，而不是一次性重写全仓库。
2. 每个模块的迁移顺序固定为：
   `Handler DTO boundary -> Usecase command/query/result -> Domain entity -> Repository adapter -> Legacy GORM model/repo`
3. 在迁移过程中允许顶层装配层暂时桥接新旧模块，但新模块内部必须符合 Clean Architecture 约束。
4. 外部兼容性默认保持不变：
   - 不改 HTTP 路由。
   - 不改 JSON 字段。
   - 不改状态码和统一响应包裹。
   - 不改既有迁移链路与数据库事实源。
5. 新模块完成迁移后，应将其作为后续模块的复制模板，而不是重新发明另一套边界方案。

## 结果

1. `user` 已成为后续 `auth`、`rbac` 重构的标准样板。
2. 架构演进从“全局大改”转为“局部可验证演进”，可以在每一轮结束后通过测试立即验收。
3. Delivery 与 Adapter 的职责边界更清晰：
   - Handler 负责 DTO 边界。
   - Repository adapter 负责 Domain 与 GORM Model 的转换。

## 守卫规则

- 不得以“先跑通”为由让新模块重新直接依赖 HTTP DTO、GORM Model 或基础设施类型。
- 不得在没有明确 ADR 的情况下破坏对外 API 契约。
- 若模块尚未完全迁移，允许通过 adapter 做过渡，但过渡层必须是单向收敛，而不是引入新的混合层。
