package sample

import (
	"context"
	"strings"
	"testing"
)

func TestDefaultToolkitDemosBuildPreviews(t *testing.T) {
	t.Parallel()

	for _, demo := range DefaultToolkitDemos() {
		demo := demo
		t.Run(demo.Module(), func(t *testing.T) {
			t.Parallel()

			result, err := demo.Build(context.Background())
			if err != nil {
				t.Fatalf("Build() returned error: %v", err)
			}
			if result.Module == "" {
				t.Fatal("Module should not be empty")
			}
			if result.Guidance == "" {
				t.Fatal("Guidance should not be empty")
			}
			if strings.TrimSpace(result.Preview) == "" {
				t.Fatal("Preview should not be empty")
			}
		})
	}
}
