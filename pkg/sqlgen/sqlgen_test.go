package sqlgen

import (
	"strings"
	"testing"
	"time"
)

// ============================================================================
// 测试模型
// ============================================================================

type TestUser struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	Username  string     `gorm:"column:username;size:64;not null"`
	Email     string     `gorm:"column:email;size:128"`
	Status    int        `gorm:"column:status;default:1"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (TestUser) TableName() string {
	return "users"
}

// ============================================================================
// Generator 测试
// ============================================================================

func TestNew(t *testing.T) {
	gen := New(nil)
	if gen == nil {
		t.Error("New() returned nil")
	}

	gen = New(&Config{Dialect: PostgreSQL})
	if gen.config.Dialect != PostgreSQL {
		t.Error("Dialect not set correctly")
	}
}

// ============================================================================
// DDL 测试
// ============================================================================

func TestTable(t *testing.T) {
	gen := New(&Config{Dialect: MySQL})

	sql, err := gen.Table(&TestUser{})
	if err != nil {
		t.Fatalf("Table() failed: %v", err)
	}

	if !strings.Contains(sql, "CREATE TABLE") {
		t.Error("SQL should contain CREATE TABLE")
	}

	if !strings.Contains(sql, "`users`") {
		t.Error("SQL should contain table name")
	}
}

func TestDrop(t *testing.T) {
	gen := New(&Config{Dialect: MySQL})

	sql, err := gen.Drop(&TestUser{})
	if err != nil {
		t.Fatalf("Drop() failed: %v", err)
	}

	expected := "DROP TABLE IF EXISTS `users`;"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

func TestDDLFromSchema(t *testing.T) {
	gen := New(&Config{Dialect: MySQL})

	schema := &Schema{
		TableName: "products",
		Fields: []Field{
			{Column: Column{Name: "id", Type: "BIGINT UNSIGNED", PrimaryKey: true, NotNull: true, AutoIncrement: true}},
			{Column: Column{Name: "name", Type: "VARCHAR(128)", NotNull: true}},
			{Column: Column{Name: "price", Type: "DECIMAL(10,2)", NotNull: false}},
			{Column: Column{Name: "deleted_at", Type: "DATETIME", NotNull: false}},
		},
		Indexes: []Index{
			{Name: "uk_products_name", Columns: []string{"name"}, Unique: true},
		},
	}

	sql, err := gen.DDLFromSchema(schema, true)
	if err != nil {
		t.Fatalf("DDLFromSchema() failed: %v", err)
	}

	if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS") {
		t.Error("SQL should contain CREATE TABLE IF NOT EXISTS")
	}
	if !strings.Contains(sql, "`products`") {
		t.Error("SQL should contain table name")
	}
	if !strings.Contains(sql, "CONSTRAINT") {
		t.Error("SQL should contain UNIQUE constraint")
	}
}

// ============================================================================
// toSnakeCase 测试 (bug 修复验证)
// ============================================================================

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserID", "user_id"},
		{"CreatedAt", "created_at"},
		{"XMLParser", "xml_parser"},
		{"SysUser", "sys_user"},
		{"HTTPSRequest", "https_request"},
		{"ID", "id"},
		{"UserName", "user_name"},
	}

	for _, tt := range tests {
		result := toSnakeCase(tt.input)
		if result != tt.expected {
			t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// ============================================================================
// INSERT 测试
// ============================================================================

func TestCreate(t *testing.T) {
	gen := New(&Config{Dialect: MySQL})

	user := TestUser{
		Username: "admin",
		Email:    "admin@test.com",
		Status:   1,
	}

	sql, err := gen.Create(&user)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if !strings.Contains(sql, "INSERT INTO") {
		t.Error("SQL should contain INSERT INTO")
	}

	if !strings.Contains(sql, "`users`") {
		t.Error("SQL should contain table name")
	}

	if !strings.Contains(sql, "'admin'") {
		t.Error("SQL should contain username value")
	}
}

func TestCreateBatch(t *testing.T) {
	gen := New(&Config{Dialect: MySQL})

	users := []TestUser{
		{Username: "user1", Email: "user1@test.com"},
		{Username: "user2", Email: "user2@test.com"},
	}

	sql, err := gen.Create(&users)
	if err != nil {
		t.Fatalf("Create batch failed: %v", err)
	}

	if !strings.Contains(sql, "INSERT INTO") {
		t.Error("SQL should contain INSERT INTO")
	}

	// 应该包含两组值
	if strings.Count(sql, "(") < 3 { // 列 + 2 组值
		t.Error("SQL should contain multiple value sets")
	}
}

// ============================================================================
// SELECT 测试
// ============================================================================

func TestFirst(t *testing.T) {
	gen := New(&Config{Dialect: MySQL, SoftDelete: true})

	var user TestUser
	sql, err := gen.First(&user, 1)
	if err != nil {
		t.Fatalf("First() failed: %v", err)
	}

	if !strings.Contains(sql, "SELECT") {
		t.Error("SQL should contain SELECT")
	}

	if !strings.Contains(sql, "LIMIT 1") {
		t.Error("SQL should contain LIMIT 1")
	}

	if !strings.Contains(sql, "id = 1") {
		t.Error("SQL should contain id condition")
	}
}

func TestFind(t *testing.T) {
	gen := New(&Config{Dialect: MySQL, SoftDelete: false})

	var users []TestUser
	sql, err := gen.Where("status = ?", 1).Order("created_at DESC").Limit(10).Offset(20).Find(&users)
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	if !strings.Contains(sql, "SELECT *") {
		t.Error("SQL should contain SELECT *")
	}

	if !strings.Contains(sql, "status = 1") {
		t.Error("SQL should contain status condition")
	}

	if !strings.Contains(sql, "ORDER BY created_at DESC") {
		t.Error("SQL should contain ORDER BY")
	}

	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("SQL should contain LIMIT")
	}

	if !strings.Contains(sql, "OFFSET 20") {
		t.Error("SQL should contain OFFSET")
	}
}

// ============================================================================
// UPDATE 测试
// ============================================================================

func TestUpdates(t *testing.T) {
	gen := New(&Config{Dialect: MySQL})

	sql, err := gen.Model(&TestUser{}).Where("id = ?", 1).Updates(map[string]interface{}{
		"username": "newname",
		"status":   2,
	})
	if err != nil {
		t.Fatalf("Updates() failed: %v", err)
	}

	if !strings.Contains(sql, "UPDATE") {
		t.Error("SQL should contain UPDATE")
	}

	if !strings.Contains(sql, "SET") {
		t.Error("SQL should contain SET")
	}

	if !strings.Contains(sql, "id = 1") {
		t.Error("SQL should contain WHERE condition")
	}
}

// ============================================================================
// DELETE 测试
// ============================================================================

func TestDelete(t *testing.T) {
	gen := New(&Config{Dialect: MySQL, SoftDelete: true})

	sql, err := gen.Delete(&TestUser{}, 1)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// 软删除应该是 UPDATE
	if !strings.Contains(sql, "UPDATE") {
		t.Error("Soft delete should use UPDATE")
	}

	if !strings.Contains(sql, "deleted_at") {
		t.Error("Soft delete should set deleted_at")
	}
}

func TestHardDelete(t *testing.T) {
	gen := New(&Config{Dialect: MySQL, SoftDelete: true})

	sql, err := gen.Unscoped().Delete(&TestUser{}, 1)
	if err != nil {
		t.Fatalf("Unscoped Delete() failed: %v", err)
	}

	if !strings.Contains(sql, "DELETE FROM") {
		t.Error("Hard delete should use DELETE FROM")
	}
}

// ============================================================================
// 逆向生成测试
// ============================================================================

func TestParseSQL(t *testing.T) {
	gen := New(&Config{Dialect: MySQL})

	ddl := `
	CREATE TABLE sys_users (
		id bigint unsigned AUTO_INCREMENT PRIMARY KEY,
		username varchar(64) NOT NULL COMMENT '用户名',
		is_active tinyint(1) DEFAULT 1,
		created_at datetime
	);`

	code, err := gen.ParseSQL(ddl).
		Package("models").
		Tags(TagGorm | TagJson).
		WithTableName(true).
		Generate()

	if err != nil {
		t.Fatalf("ParseSQL().Generate() failed: %v", err)
	}

	if !strings.Contains(code, "package models") {
		t.Error("Code should contain package declaration")
	}

	if !strings.Contains(code, "type SysUsers struct") {
		t.Error("Code should contain struct definition")
	}

	if !strings.Contains(code, "gorm:") {
		t.Error("Code should contain gorm tags")
	}

	if !strings.Contains(code, "json:") {
		t.Error("Code should contain json tags")
	}

	if !strings.Contains(code, "TableName()") {
		t.Error("Code should contain TableName method")
	}
}

// ============================================================================
// 方言测试
// ============================================================================

func TestDialectQuote(t *testing.T) {
	tests := []struct {
		dialect  Dialect
		input    string
		expected string
	}{
		{MySQL, "users", "`users`"},
		{PostgreSQL, "users", "\"users\""},
		{SQLite, "users", "\"users\""},
		{SQLServer, "users", "[users]"},
	}

	for _, tt := range tests {
		d := getDialect(tt.dialect)
		result := d.Quote(tt.input)
		if result != tt.expected {
			t.Errorf("Dialect %s: Quote(%q) = %q, expected %q",
				tt.dialect, tt.input, result, tt.expected)
		}
	}
}

// ============================================================================
// 事务测试
// ============================================================================

func TestTransaction(t *testing.T) {
	gen := New(&Config{Dialect: MySQL})

	sql := gen.Transaction(func(tx *Generator) string {
		s1, _ := tx.Create(&TestUser{Username: "user1"})
		return s1
	})

	if !strings.Contains(sql, "START TRANSACTION") {
		t.Error("Transaction should start with START TRANSACTION")
	}

	if !strings.Contains(sql, "COMMIT") {
		t.Error("Transaction should end with COMMIT")
	}

	if !strings.Contains(sql, "INSERT INTO") {
		t.Error("Transaction should contain INSERT")
	}
}

// ============================================================================
// RenderScript 测试
// ============================================================================

func TestRenderScript(t *testing.T) {
	stmts := []string{"SELECT 1;", "SELECT 2;"}
	result := RenderScript(stmts)

	if !strings.Contains(result, "SELECT 1;") {
		t.Error("RenderScript should contain first statement")
	}
	if !strings.Contains(result, "SELECT 2;") {
		t.Error("RenderScript should contain second statement")
	}

	if RenderScript(nil) != "" {
		t.Error("RenderScript with empty input should return empty string")
	}
}
