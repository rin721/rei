package sample

import (
	"context"
	"strings"

	"github.com/rin721/rei/pkg/sqlgen"
	"github.com/rin721/rei/pkg/yaml2go"
	"github.com/rin721/rei/types"
)

// ToolkitDemo describes a DI-friendly sample use case for an isolated pkg helper.
type ToolkitDemo interface {
	Module() string
	Build(context.Context) (types.SampleToolkitDemoResponse, error)
}

func DefaultToolkitDemos() []ToolkitDemo {
	return []ToolkitDemo{
		sqlgenToolkitDemo{},
		yaml2goToolkitDemo{},
	}
}

type sqlgenToolkitDemo struct{}

func (sqlgenToolkitDemo) Module() string {
	return "sqlgen"
}

// TODO: SQLGen Demo - use sqlgen to preview forward and reverse schema changes
// before checking versioned SQL into scripts/migrations.
func (sqlgenToolkitDemo) Build(_ context.Context) (types.SampleToolkitDemoResponse, error) {
	generator := sqlgen.New(&sqlgen.Config{
		Dialect: sqlgen.SQLite,
		Pretty:  true,
	})

	ddl, err := generator.Table(sampleIntegrationRecord{})
	if err != nil {
		return types.SampleToolkitDemoResponse{}, err
	}

	reverse, err := generator.
		ParseSQL(ddl).
		Package("generated").
		Name("SampleIntegrationRecord").
		Generate()
	if err != nil {
		return types.SampleToolkitDemoResponse{}, err
	}

	return types.SampleToolkitDemoResponse{
		Module:   "sqlgen",
		Scenario: "Offline migration preview for a new integration table",
		Guidance: "Generate and review SQL offline first, then persist the approved statement as a versioned migration under scripts/migrations.",
		Preview:  strings.TrimSpace(ddl) + "\n\n" + strings.TrimSpace(reverse),
	}, nil
}

type yaml2goToolkitDemo struct{}

func (yaml2goToolkitDemo) Module() string {
	return "yaml2go"
}

// TODO: YAML2Go Demo - use yaml2go to scaffold typed config contracts from
// external YAML payloads before promoting them into stable module configs.
func (yaml2goToolkitDemo) Build(_ context.Context) (types.SampleToolkitDemoResponse, error) {
	source := []byte(
		"integration:\n" +
			"  name: billing\n" +
			"  enabled: true\n" +
			"  retries: 3\n" +
			"  endpoints:\n" +
			"    - path: /invoices\n" +
			"      method: POST\n",
	)

	code, err := yaml2go.Convert("PartnerIntegrationConfig", source)
	if err != nil {
		return types.SampleToolkitDemoResponse{}, err
	}

	return types.SampleToolkitDemoResponse{
		Module:   "yaml2go",
		Scenario: "Scaffold Go structs from a partner YAML contract",
		Guidance: "Use the generated struct as a starting point, then move the stabilized config type into the owning module after field names and tags are reviewed.",
		Preview:  strings.TrimSpace(code),
	}, nil
}

type sampleIntegrationRecord struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	ExternalRef string `gorm:"column:external_ref;type:varchar(64);not null;uniqueIndex"`
	Status      string `gorm:"column:status;type:varchar(32);not null;index"`
}

func (sampleIntegrationRecord) TableName() string {
	return "sample_integration_records"
}
