# sqlgen 包概述

> **sqlgen** 是一个"离线版 GORM"——它不连接数据库，而是以纯文本方式生成、解析 SQL，为脚手架工具、代码生成器和 CI 流水线提供数据库层能力。

---

## 定位与设计哲学

| 设计原则 | 说明 |
|---------|------|
| **纯文本工具** | 零数据库连接依赖，可在任意环境运行（CI、代码生成器、离线工具） |
| **GORM 兼容** | 链式 API 与 GORM 高度一致，降低学习成本 |
| **双向生成** | 正向：Go Struct → SQL；逆向：SQL DDL → Go Struct |
| **多方言** | 原生支持 MySQL、PostgreSQL、SQLite、SQL Server |
| **可扩展性** | 支持自定义方言、类型映射、模板、Hooks、字段转换器 |

---

## 目录结构

```
pkg/sqlgen/
├── doc.go           # 包文档，快速开始示例
├── types.go         # 所有公共类型定义 (Config, Schema, Field, Column, …)
├── constants.go     # 常量与枚举 (Dialect, TagType, NamingStrategy, OperationType)
├── errors.go        # 错误类型、错误码、预定义错误变量
├── generator.go     # Generator 主结构、链式调用方法、命名工具函数
├── dialect.go       # DialectHandler 接口 + 4 种方言实现 + SQL 插值
├── reflect.go       # 反射解析 Go Struct → FieldInfo，读取 gorm tag
├── create.go        # INSERT 生成 (单行 / 批量 / UPSERT / Truncate)
├── query.go         # SELECT / COUNT / First / Find 生成
├── update.go        # UPDATE / Save / UpdateColumn 生成
├── delete.go        # DELETE（含软删除）生成
├── ddl.go           # CREATE TABLE / DROP TABLE / ALTER TABLE / MigrateBuilder
├── transaction.go   # 事务脚本 / BatchBuilder / MigrationBuilder / SeedBuilder
├── parser.go        # SQL DDL 解析器 (CREATE TABLE → Schema)
├── reverse.go       # 逆向生成入口：ReverseBuilder / DBReverseBuilder
├── codegen.go       # Go 代码生成器 (Struct + DAO)
├── template.go      # Go text/template 封装，DefaultStructTemplate / DefaultDAOTemplate
└── sqlgen_test.go   # 完整测试用例
```

---

## 核心类型

### Generator（主结构）

```go
type Generator struct {
    config  *Config          // 全局配置
    dialect DialectHandler   // 方言处理器
    ctx     *QueryContext    // 链式调用上下文（每次克隆）
    mu      sync.RWMutex
}
```

- 通过 `New(cfg *Config)` 创建，`nil` 时使用 `DefaultConfig()`
- 每个链式方法（`Where`、`Order`、`Limit` 等）返回**新克隆的** Generator，保持不可变性，线程安全

### Config

```go
type Config struct {
    Dialect             Dialect   // 数据库方言
    Pretty              bool      // 格式化输出
    SkipZeroValue       bool      // UPDATE 跳过零值字段
    SoftDelete          bool      // 自动检测 deleted_at 软删除
    AllowEmptyCondition bool      // 允许无条件 UPDATE/DELETE
}
```

### Schema（DDL 解析结果）

```go
type Schema struct {
    Name      string   // 结构体名（PascalCase）
    TableName string   // 数据库表名（snake_case）
    Fields    []Field  // 字段列表
    Comment   string   // 表注释
    Indexes   []Index  // 索引列表
    Package   string   // 生成代码包名
    Imports   []string // 需要导入的包
}
```

### ReverseOptions（逆向生成配置）

提供完整的代码生成控制：命名策略（`FieldNaming`、`JSONNaming`、`FileNaming`）、Tag 标志位（`Tags`）、开关（`WithComments`、`WithTableName`、`WithSoftDelete`）、钩子（`BeforeGenerate`、`AfterGenerate`）、自定义类型映射（`TypeMappings`）、字段转换器（`FieldConverter`）。

---

## 功能模块

### 1. 正向生成（Go Struct → SQL）

#### CRUD

| 方法 | 生成 SQL | 说明 |
|------|---------|------|
| `Create(model)` | `INSERT INTO …` | 支持单行和批量（slice），自动跳过零值和自增列 |
| `Find(model)` | `SELECT * FROM …` | 支持 Where/Order/Limit/Offset/Select/Omit |
| `First(model, id)` | `SELECT … LIMIT 1` | 自动追加主键条件 + `LIMIT 1` |
| `Count(model)` | `SELECT COUNT(*) …` | 支持 WHERE 条件 |
| `Updates(map/struct)` | `UPDATE … SET …` | 无 WHERE 条件时返回错误（可配置关闭保护） |
| `Save(model)` | `UPDATE … SET …` | 更新所有字段 |
| `UpdateColumn(col, val)` | `UPDATE … SET col=val` | 单列更新 |
| `Delete(model, id)` | 软删除 → `UPDATE set deleted_at=NOW()` / 硬删除 → `DELETE FROM` | 由 `SoftDelete` 配置控制 |

#### 链式条件 API

```go
gen.
    Where("status = ?", 1).
    Where("created_at > ?", "2024-01-01").
    Order("created_at DESC").
    Limit(20).
    Offset(40).
    Select("id", "username").
    Find(&users)
```

#### DDL

| 方法 | 生成 SQL |
|------|---------|
| `Table(model)` | `CREATE TABLE …`（反射 gorm tag） |
| `Drop(model)` | `DROP TABLE IF EXISTS …` |
| `DDLFromSchema(schema, ifNotExists)` | 从 `Schema` 结构体生成 CREATE TABLE，支持 UNIQUE 约束 |
| `Migrate(model)` | 返回 `MigrateBuilder`，支持 AddColumn / DropColumn / ModifyColumn / AddIndex / DropIndex |

#### 事务与脚本

| 方法/类型 | 说明 |
|---------|------|
| `Transaction(fn)` | 生成 `START TRANSACTION … COMMIT` 脚本，自动适配方言 |
| `TransactionWithRollback(fn, rollbackFn)` | 同时输出回滚脚本（注释形式） |
| `BatchBuilder` | 收集多条语句，`Build()` 或 `BuildWithTransaction()` 输出 |
| `MigrationBuilder` | 按时间戳命名的迁移脚本，支持 `Build()` 正向和 `BuildRollback()` 回滚 |
| `SeedBuilder` | 种子数据脚本，可选先 TRUNCATE 再 INSERT |

#### 原生 SQL

```go
sql, _ := gen.Raw("SELECT * FROM users WHERE id = ?", 42).Build()
```

---

### 2. 逆向生成（SQL DDL → Go Struct）

**入口**：`gen.ParseSQL(ddl)` / `gen.ParseSQLFile(path)` → 返回 `ReverseBuilder`

#### ReverseBuilder 链式 API

```go
gen.ParseSQL(ddl).
    Package("models").
    Tags(TagGorm | TagJson).
    FieldNaming(PascalCase).
    JSONTagNaming(SnakeCase).
    WithComments(true).
    WithTableName(true).
    WithSoftDelete(true).
    TypeMapping("json", "json.RawMessage").
    BeforeGenerate(func(s *Schema) { /* 修改 schema */ }).
    AfterGenerate(func(code string) string { return code }).
    Generate()
```

#### 生成方法

| 方法 | 说明 |
|------|------|
| `Generate()` | 生成第一个表的 Go Struct 字符串 |
| `GenerateAll()` | 生成所有表，返回 `map[tableName]code` |
| `GenerateToFile(path)` | 写入单个文件（含 Overwrite 保护） |
| `GenerateToDir(dir)` | 每个表一个文件，按 `FileNaming` 策略命名 |
| `GenerateWithDAO()` | 同时生成 Struct + DAO 代码 |

#### DDL 解析器（`parser.go`）

- 正则 + 手动括号平衡匹配，可正确处理嵌套括号
- 支持 `IF NOT EXISTS`、反引号/双引号/方括号标识符
- 解析列属性：类型（含参数）、`NOT NULL`、`AUTO_INCREMENT`/`AUTOINCREMENT`/`SERIAL`/`IDENTITY`、`DEFAULT`、`COMMENT`
- 解析表级 `PRIMARY KEY` 约束
- 忽略 `KEY`、`INDEX`、`UNIQUE`、`FOREIGN KEY`、`CHECK` 约束行
- 自动分析 `import`（`time`、`encoding/json`）

---

### 3. DAO 代码生成（`codegen.go`）

`CodeGenerator.GenerateDAO(schema, methods)` 支持生成以下方法：

| 方法名 | 生成签名 |
|--------|---------|
| `Create` | `func (d *XxxDAO) Create(entity *Xxx) error` |
| `Update` | `func (d *XxxDAO) Update(entity *Xxx) error` |
| `Delete` | `func (d *XxxDAO) Delete(id PkType) error` |
| `FindByID` | `func (d *XxxDAO) FindByID(id PkType) (*Xxx, error)` |
| `FindAll` | `func (d *XxxDAO) FindAll() ([]*Xxx, error)` |

自动检测主键字段类型作为 `id` 参数类型。

---

### 4. 方言系统（`dialect.go`）

`DialectHandler` 接口定义了方言的全部行为：

```go
type DialectHandler interface {
    Name() Dialect
    Quote(name string) string            // 标识符引号（`x` / "x" / [x]）
    Placeholder(index int) string        // 占位符（? / $1 / @p1）
    TypeMapping(goType string, size int) string     // Go 类型 → SQL 类型
    ReverseTypeMapping(sqlType string) string       // SQL 类型 → Go 类型
    Interpolate(sql string, args ...interface{}) (string, error)
    AutoIncrementKeyword() string
    DefaultValueKeyword() string
    EngineClause() string               // MySQL ENGINE 子句
}
```

| 方言 | 引号 | 占位符 | 自增关键字 |
|------|------|--------|----------|
| MySQL | `` ` `` | `?` | `AUTO_INCREMENT` |
| PostgreSQL | `"` | `$1` | （用 SERIAL 类型） |
| SQLite | `"` | `?` | `AUTOINCREMENT` |
| SQL Server | `[` `]` | `@p1` | `IDENTITY(1,1)` |

使用 `RegisterDialect(d, handler)` 注册自定义方言。

---

### 5. 反射系统（`reflect.go`）

- 读取 Go Struct 字段的 `gorm` tag，解析 `column`、`primaryKey`、`autoIncrement`、`not null`、`default`、`size`、`type`、`index`、`uniqueIndex`、`comment`
- 自动将 Go 类型通过 `DialectHandler.TypeMapping()` 转为 SQL 类型
- 软删除检测：字段名含 `deleted_at` 且类型为 `*time.Time`

---

### 6. 错误系统（`errors.go`）

```go
type Error struct {
    Code    ErrorCode
    Message string
    Cause   error   // 支持 errors.Unwrap 链
}
```

| 预定义错误 | 触发场景 |
|-----------|---------|
| `ErrInvalidModel` | 传入非结构体指针 |
| `ErrInvalidDialect` | 不支持的方言 |
| `ErrParseFailed` | DDL 解析失败 |
| `ErrMissingCondition` | UPDATE/DELETE 无 WHERE 且未开启 AllowEmptyCondition |
| `ErrEmptyData` | INSERT 空数据 |
| `ErrNoTableName` | 无法推断表名 |

`IsError(err, code)` 用于精确匹配错误码。

---

## 快速上手

### 正向生成

```go
gen := sqlgen.New(&sqlgen.Config{
    Dialect:    sqlgen.MySQL,
    SoftDelete: true,
})

// INSERT
sql, _ := gen.Create(&user)
// => INSERT INTO `users` (`username`, `email`) VALUES ('admin', 'admin@example.com')

// SELECT
sql, _ = gen.Where("status = ?", 1).Order("id DESC").Limit(10).Find(&users)
// => SELECT * FROM `users` WHERE status = 1 AND `deleted_at` IS NULL ORDER BY id DESC LIMIT 10

// UPDATE
sql, _ = gen.Model(&User{}).Where("id = ?", 1).Updates(map[string]interface{}{"status": 2})
// => UPDATE `users` SET `status` = 2 WHERE id = 1

// DELETE（软删除）
sql, _ = gen.Delete(&user, 1)
// => UPDATE `users` SET `deleted_at` = '…' WHERE `id` = 1

// CREATE TABLE
sql, _ = gen.Table(&user)
// => CREATE TABLE `users` (…) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 逆向生成

```go
ddl := `
CREATE TABLE sys_users (
    id bigint unsigned AUTO_INCREMENT PRIMARY KEY,
    username varchar(64) NOT NULL COMMENT '用户名',
    is_active tinyint(1) DEFAULT 1,
    created_at datetime,
    deleted_at datetime
);`

code, _ := gen.ParseSQL(ddl).
    Package("models").
    Tags(sqlgen.TagGorm | sqlgen.TagJson).
    WithTableName(true).
    Generate()

// 输出:
// package models
//
// import "time"
//
// // SysUsers sys_users
// type SysUsers struct {
//     ID        uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
//     Username  string     `gorm:"column:username;type:varchar(64);not null;comment:用户名" json:"username"`
//     IsActive  bool       `gorm:"column:is_active;type:tinyint(1);default:1" json:"is_active"`
//     CreatedAt time.Time  `gorm:"column:created_at;type:datetime" json:"created_at"`
//     DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime" json:"deleted_at"`
// }
//
// func (SysUsers) TableName() string { return "sys_users" }
```

### 事务脚本

```go
script := gen.Transaction(func(tx *sqlgen.Generator) string {
    s1, _ := tx.Create(&User{Username: "admin"})
    s2, _ := tx.Model(&User{}).Where("id = ?", 1).Updates(map[string]interface{}{"status": 1})
    return s1 + "\n" + s2
})
// => START TRANSACTION;
//    INSERT INTO `users` …;
//    UPDATE `users` SET …;
//    COMMIT;
```

---

## 常量速查

### TagType（位标志，可组合）

| 常量 | 值 | 说明 |
|------|-----|------|
| `TagGorm` | 1 | 生成 `gorm` tag |
| `TagJson` | 2 | 生成 `json` tag |
| `TagXml` | 4 | 生成 `xml` tag |
| `TagYaml` | 8 | 生成 `yaml` tag |
| `TagValidate` | 16 | 生成 `validate:"required"` tag |
| `TagDefault` | 3 | TagGorm \| TagJson |
| `TagAll` | 31 | 所有 tag |

### NamingStrategy

| 常量 | 示例 |
|------|------|
| `SnakeCase` | `user_name` |
| `CamelCase` | `userName` |
| `PascalCase` | `UserName` |
| `KebabCase` | `user-name` |

---

## 已知限制 / TODO

- `DBReverseBuilder`（从数据库连接直接逆向）尚未实现，调用 `Generate()` 等方法会返回 `ErrCodeUnknown`
- `MergeWithFile` 增量合并接口已声明但逻辑未完整实现
- `MigrationBuilder.ModifyColumn` 回滚时需要原始类型，当前输出为 `-- TODO` 注释
- PostgreSQL `BIGSERIAL` 自增的逆向类型映射为 `int64`（正确），但正向 DDL 生成中 PostgreSQL 使用 SERIAL 类型而非 `AUTO_INCREMENT` 关键字（正确处理，关键字返回空字符串）

---

## 与项目其他模块的集成

`sqlgen` 在整个脚手架项目中作为**数据库 DDL 迁移脚本生成器**使用：

- **`pkg/migrate/generator.go`**：调用 `gen.Table()` 和 `gen.Drop()` 生成版本化迁移脚本
- **`scripts/migrations/*.sql`**：作为数据库结构历史的版本化输出
