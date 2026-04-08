package sqlgen

import "testing"

func TestGenerateAndParseCreateTable(t *testing.T) {
	t.Parallel()

	statement, err := GenerateCreateTableWithOptions(Table{
		Name: "users",
		Columns: []Column{
			{Name: "id", Type: "integer", PrimaryKey: true, Nullable: false},
			{Name: "email", Type: "text", Nullable: false, Unique: true},
		},
		UniqueConstraints: []UniqueConstraint{
			{Columns: []string{"id", "email"}},
		},
	}, GenerateOptions{
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("GenerateCreateTableWithOptions() returned error: %v", err)
	}

	table, err := ParseCreateTable(statement)
	if err != nil {
		t.Fatalf("ParseCreateTable() returned error: %v", err)
	}
	if table.Name != "users" {
		t.Fatalf("table.Name = %q, want %q", table.Name, "users")
	}
	if len(table.Columns) != 2 {
		t.Fatalf("len(table.Columns) = %d, want 2", len(table.Columns))
	}
	if !table.Columns[1].Unique {
		t.Fatal("table.Columns[1].Unique = false, want true")
	}
}

func TestGenerateStatementsAndRenderScript(t *testing.T) {
	t.Parallel()

	statements, err := GenerateStatements([]Table{
		{
			Name: "roles",
			Columns: []Column{
				{Name: "id", Type: "integer", PrimaryKey: true, Nullable: false},
				{Name: "name", Type: "varchar(64)", Nullable: false, Unique: true},
			},
		},
		{
			Name: "samples",
			Columns: []Column{
				{Name: "id", Type: "integer", PrimaryKey: true, Nullable: false},
				{Name: "enabled", Type: "boolean", Nullable: false, Default: "true"},
			},
		},
	}, GenerateOptions{
		IfNotExists: true,
	})
	if err != nil {
		t.Fatalf("GenerateStatements() returned error: %v", err)
	}
	if len(statements) != 2 {
		t.Fatalf("len(statements) = %d, want 2", len(statements))
	}

	script := RenderScript(statements)
	if script == "" {
		t.Fatal("RenderScript() returned empty script")
	}
}
