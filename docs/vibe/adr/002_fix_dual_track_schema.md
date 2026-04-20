# ADR 002: Cut Legacy InitDB and Unify Schema Source

日期：2026-04-20

## 背景

仓库曾同时存在旧版 `initdb` 链路、运行时 `AutoMigrate` 以及新的 `cmd/db + pkg/migrate` 迁移体系，导致 Schema 来源分裂、测试与运行时行为不一致。

## 决策

1. 删除旧版 `cmd/initdb` 与 `internal/app/app_initdb*` 链路。
2. 删除 `ModeInitDB`，仅保留 `ModeDB` 作为数据库管理模式。
3. 禁止在运行时和测试中直接使用 `AutoMigrate`。
4. 确立以下唯一建库标准：
   `database.driver + database.migrations_dir + cmd/db + pkg/migrate + scripts/migrations`

## 结果

1. 服务启动不再承担 Schema 变更职责。
2. 测试通过真实迁移脚本准备数据库结构。
3. 配置面不再保留 `initdb` 命名与 lock-file 机制。
4. 仓库对外文档与示例配置统一指向迁移工作流。
