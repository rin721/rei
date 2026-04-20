# ADR 003: Runtime and Tests Share the Same Migration Artifacts

日期: 2026-04-20

状态: Accepted

## 背景

在切除旧版 `initdb` 双轨制之后，仓库已经明确了 `cmd/db + pkg/migrate + scripts/migrations` 是唯一的 Schema 事实源。

这次重构过程中又形成了一条更细的工程共识：不仅生产运行时不能再偷偷创建或修正表结构，测试环境也必须复用同一套迁移产物。否则，测试通过的结构与真实运行结构仍然可能再次分叉。

## 决策

1. `ModeServer` 运行时不得承担 Schema 变更职责。
2. 所有需要准备数据库结构的测试，必须通过 `pkg/migrate` 执行 `scripts/migrations` 中的真实迁移脚本。
3. `dry-run`、`status` 等只读或预演行为不得创建业务表，也不得偷偷写入 `schema_migrations`。
4. 迁移历史表的维护必须使用显式 DDL，不得重新引入 `AutoMigrate` 作为兜底实现。
5. 迁移配置统一归入 `database.migrations_dir`，不再保留独立的 `initdb` 配置命名空间。

## 结果

1. 运行时、测试、CLI 三条链路消费同一套迁移资产，结构来源完全一致。
2. 测试不再依赖 ORM 推断出的临时表结构，从而更接近真实部署行为。
3. 迁移命令的副作用边界更清晰，预演模式可以安全用于验证。
4. 后续任何表、索引、约束变更都必须首先体现在迁移脚本中，而不是散落在启动逻辑或测试辅助代码里。

## 守卫规则

- 禁止在应用启动路径重新加入 `AutoMigrate`。
- 禁止在测试中通过 ORM 自动建表替代真实迁移。
- 禁止新增第二套 Schema 生成或落库入口。
- 如果未来需要扩展迁移能力，应继续围绕 `pkg/migrate` 演进，而不是恢复独立初始化工具。
