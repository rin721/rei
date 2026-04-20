# Current Task

Status: Completed

Completed On: 2026-04-20

Completed Task: 彻底切除废弃的旧版 `initdb` 链路，并统一 `cmd/db + pkg/migrate + scripts/migrations` 为唯一 Schema 事实源。

Completion Notes:
- 已物理删除旧 `initdb` CLI、`internal/app` 初始化链路与历史脚本目录。
- 已移除运行时与测试中的直接 `AutoMigrate` 调用。
- 已将测试建库流程统一为通过真实迁移脚本准备 Schema。
