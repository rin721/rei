# Agent Memory Base

本系统采用严格的 Clean Architecture 与 DDD。

所有新增开发、重构、评审、脚手架演进与模块扩展，必须优先遵守 `docs/vibe/constitution/` 目录中的强制规范。

执行顺序要求如下：

1. 先查阅 `constitution/clean_architecture.md`，确认分层边界与依赖方向。
2. 再查阅 `constitution/tech_stack.md`，确认技术栈约束、迁移入口与禁用链路。
3. 如遇到当前重构上下文，优先参考 `state/CURRENT_TASK.md`。
4. 如需理解历史决策，查阅 `adr/` 目录中的记录。

本目录是 Agent 的长期记忆底座，任何实现都不得绕过这里定义的工程约束。
