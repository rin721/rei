# pkg/migrate 包概述

> **migrate** 是项目的数据库迁移管理层，提供版本化脚本生成、执行、状态查看与回滚能力，支持多实例部署与完整的回滚流程。

---

## 定位与设计哲学

| 设计原则 | 说明 |
|---------|------|
| **两阶段分离** | Generate（离线，无需 DB）→ Migrate（在线执行），职责清晰 |
| **数据库锁表** | `schema_migrations` 替代文件锁，天然支持多实例并发 |
| **校验和保护** | 已执行迁移若 up 文件被篡改，立即报错拒绝执行 |
| **幂等安全** | 所有 DDL 使用 `CREATE TABLE IF NOT EXISTS` |
| **版本化** | `YYYYMMDD_NNN_desc.up/down.sql` 按时间+序号排序 |
| **Model 驱动** | Schema 来自 Go Model struct 反射，与业务代码保持同步 |

---

## 目录结构

```
pkg/migrate/
├── types.go      # 公共类型定义（Migration, MigrationRecord, Status…）
├── history.go    # schema_migrations 锁表读写
├── scanner.go    # 扫描 migrations/ 目录，解析版本文件
├── generator.go  # 离线生成器：反射 Go Model → 版本化 up/down SQL
└── migrate.go    # Migrator 主结构：Migrate / Rollback / Status
```

---

## 核心类型

### Migration（版本描述）

```go
type Migration struct {
    Version     string   // 排序键，格式 YYYYMMDD_NNN（例如 20260420_001）
    Description string   // 来自文件名（例如 init_schema）
    UpFile      string   // .up.sql 文件绝对路径
    DownFile    string   // .down.sql 文件绝对路径（可选）
    Checksum    string   // up 文件内容的 SHA256 十六进制字符串
}
```

### MigrationRecord（数据库锁表记录）

```go
type MigrationRecord struct {
    Version     string    // 主键，对应 Migration.Version
    Description string
    AppliedAt   time.Time
    Checksum    string    // 执行时记录的 checksum，用于后续校验
}

func (MigrationRecord) TableName() string { return "schema_migrations" }
```

### Migrator（主结构）

```go
type Migrator struct {
    db         *gorm.DB
    dialect    string
    scriptsDir string   // migrations/ 目录路径
    history    *History
}

func New(db *gorm.DB, dialect, scriptsDir string) *Migrator
```

### GenerateOptions（离线生成配置）

```go
type GenerateOptions struct {
    Models      []any    // Go Model 指针列表（例如 []any{&User{}, &Role{}}）
    OutputDir   string   // 输出目录（默认 scripts/migrations）
    Version     string   // 版本号（留空自动生成 YYYYMMDD_NNN）
    Description string   // 描述（用于文件名）
    Dialect     string   // 数据库方言
    WithCRUD    bool     // 是否同时生成 CRUD SQL 参考文档
}
```

---

## 功能模块

### 1. 离线生成（`generator.go`）

`Generate(opts GenerateOptions) error`

- 无需数据库连接，纯文件操作
- 调用 `sqlgen.New(...).Table(model)` 反射 Go Model 生成 DDL
- 自动为 DDL 添加 `IF NOT EXISTS`（幂等安全）
- SQLite 方言下自动过滤内联 `INDEX` 语法（SQLite `CREATE TABLE` 不支持）
- 版本号自动推导：当天已有文件最大序号 +1
- 生成文件对：`{version}_{desc}.up.sql` + `{version}_{desc}.down.sql`
- 可选 `WithCRUD=true`：生成 CRUD SQL 参考文档到 `scripts/crud/`

**文件命名规范：**

```
scripts/migrations/
├── 20260420_001_init_schema.up.sql
├── 20260420_001_init_schema.down.sql
├── 20260421_001_add_user_avatar.up.sql
└── 20260421_001_add_user_avatar.down.sql
```

---

### 2. 版本扫描（`scanner.go`）

`Scan(dir string) ([]*Migration, error)`

- 读取 `migrations/` 目录，按正则 `{YYYYMMDD}_{NNN}_{desc}.(up|down).sql` 匹配文件
- 自动计算 up 文件 SHA256 checksum
- 返回按 `Version` 字典序排序的 `[]*Migration`

---

### 3. 数据库锁表（`history.go`）

`History` 结构管理 `schema_migrations` 表：

| 方法 | 说明 |
|------|------|
| `EnsureTable(ctx)` | 使用显式 DDL 确保表存在（首次运行自动创建） |
| `IsApplied(ctx, version)` | 查询是否已执行 |
| `MarkApplied(ctx, rec)` | 插入一条执行记录 |
| `Unmark(ctx, version)` | 删除执行记录（回滚时调用） |
| `ListApplied(ctx)` | 返回所有已执行记录（按 applied_at 升序） |
| `LastApplied(ctx)` | 返回最近一条执行记录 |

---

### 4. Migrator 主体（`migrate.go`）

#### Migrate

```go
func (m *Migrator) Migrate(ctx context.Context, dryRun bool) ([]string, error)
```

执行流程：
1. `EnsureTable`：确保锁表存在
2. `Scan`：扫描脚本目录
3. 遍历所有 Migration：
   - 已执行 → 校验 checksum（防篡改）
   - 未执行 → 在单事务内执行 up SQL，并写入锁表记录
4. 返回本次执行的版本列表

#### Rollback

```go
func (m *Migrator) Rollback(ctx context.Context, steps int, dryRun bool) ([]string, error)
```

- 查询最近 `steps` 条锁表记录（按 applied_at DESC）
- 在单事务内执行对应 `.down.sql`，并删除锁表记录
- 若无 `.down.sql` 文件则报错退出

#### Status

```go
func (m *Migrator) Status(ctx context.Context) (*MigrationStatus, error)
```

- 返回 `Applied`（已执行版本列表）和 `Pending`（待执行版本列表）

---

## CLI 命令

```
rei db generate   --desc <desc> [--version <version>] [--with-crud] [--dry-run]
rei db migrate    [--yes] [--dry-run]
rei db status
rei db rollback   [--steps N] [--yes] [--dry-run]
```

### 典型工作流

```bash
# 1. 新增/修改 Go Model 后，离线生成迁移脚本
rei db generate --desc init_schema

# 2. 查看状态（此时显示 pending）
rei db status

# 3. 执行迁移
rei db migrate --yes

# 4. 再次查看（显示 applied）
rei db status

# 5. 如需撤销
rei db rollback --yes
```

---

## 与项目其他模块的集成

```
internal/
├── models/registry.go       # models.All()：提供 Model 列表给 db generate
└── app/
    ├── app_mode.go           # ModeDB 模式
    └── app_mode_db.go        # DB 模式入口：generate/migrate/status/rollback 分发

internal/config/
└── app_database.go           # DatabaseConfig.MigrationsDir（默认 scripts/migrations）

cmd/
└── db/
    ├── db.go                 # db 父命令组
    ├── generate.go           # rei db generate
    ├── migrate.go            # rei db migrate
    ├── status.go             # rei db status
    └── rollback.go           # rei db rollback
```

---

## 与旧建库链路的关系

| 维度 | 旧建库链路 | 新 `rei db migrate` |
|------|----------------|---------------------|
| Schema 来源 | 手动维护的独立 schema 副本 | Go Model 反射（`models.All()`） |
| 锁机制 | 本地 lock file | 数据库表 `schema_migrations` |
| 多实例支持 | ❌ | ✅ |
| 版本追踪 | ❌ | ✅（按版本号顺序） |
| 回滚能力 | ❌ | ✅（`.down.sql`） |
| 校验和保护 | ❌ | ✅（SHA256） |
| 离线生成 | ❌（生成+执行一步） | ✅（generate 完全离线） |

> 旧建库命令已移除，仓库统一使用 `rei db migrate`。

---

## 已知限制 / TODO

- SQLite 下索引（非 UNIQUE 约束）被过滤掉，未单独生成 `CREATE INDEX` 语句；如需索引需手动补充
- `DBReverseBuilder`（从数据库直接逆向生成迁移）未实现
- 当前不支持并发安全的分布式锁（同一时刻两个实例同时 migrate 时存在竞争），建议在 CI/CD 中串行执行
