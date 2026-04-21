package app

import (
	"context"
	"testing"
)

func TestInfrastructureCapabilityRegistryBootstrapOrdersDependencies(t *testing.T) {
	t.Parallel()

	registry := newInfrastructureCapabilityRegistry()
	capabilities, err := registry.orderedCapabilities(infrastructureProfileServerBootstrap, infrastructureCapabilityHookBootstrap)
	if err != nil {
		t.Fatalf("orderedCapabilities() returned error: %v", err)
	}

	positions := make(map[infrastructureCapabilityName]int, len(capabilities))
	for index, capability := range capabilities {
		positions[capability.name] = index
	}

	if positions[infrastructureCapabilityDatabase] >= positions[infrastructureCapabilityDBTx] {
		t.Fatalf("database position %d should be before dbtx position %d", positions[infrastructureCapabilityDatabase], positions[infrastructureCapabilityDBTx])
	}
}

func TestInfrastructureCapabilityRegistryShutdownReversesDependencies(t *testing.T) {
	t.Parallel()

	registry := infrastructureCapabilityRegistry{
		capabilities: map[infrastructureCapabilityName]infrastructureCapability{
			"database": {
				name:     "database",
				label:    "database",
				shutdown: func(_ context.Context, _ infrastructureProvisioning) error { return nil },
			},
			"dbtx": {
				name:                 "dbtx",
				label:                "dbtx",
				shutdownDependencies: []infrastructureCapabilityName{"database"},
				shutdown:             func(_ context.Context, _ infrastructureProvisioning) error { return nil },
			},
		},
		profiles: map[infrastructureCapabilityProfile][]infrastructureCapabilityName{
			infrastructureProfileRuntimeShutdown: {"database", "dbtx"},
		},
	}

	capabilities, err := registry.orderedCapabilities(infrastructureProfileRuntimeShutdown, infrastructureCapabilityHookShutdown)
	if err != nil {
		t.Fatalf("orderedCapabilities() returned error: %v", err)
	}
	if len(capabilities) != 2 {
		t.Fatalf("capabilities len = %d, want 2", len(capabilities))
	}
	if capabilities[0].name != "dbtx" || capabilities[1].name != "database" {
		t.Fatalf("shutdown order = [%s %s], want [dbtx database]", capabilities[0].name, capabilities[1].name)
	}
}

func TestInfrastructureCapabilityRegistryDetectsCycles(t *testing.T) {
	t.Parallel()

	registry := infrastructureCapabilityRegistry{
		capabilities: map[infrastructureCapabilityName]infrastructureCapability{
			"a": {
				name:                  "a",
				label:                 "a",
				bootstrapDependencies: []infrastructureCapabilityName{"b"},
				bootstrap:             func(_ context.Context, _ infrastructureProvisioning) error { return nil },
			},
			"b": {
				name:                  "b",
				label:                 "b",
				bootstrapDependencies: []infrastructureCapabilityName{"a"},
				bootstrap:             func(_ context.Context, _ infrastructureProvisioning) error { return nil },
			},
		},
		profiles: map[infrastructureCapabilityProfile][]infrastructureCapabilityName{
			infrastructureProfileServerBootstrap: {"a", "b"},
		},
	}

	if _, err := registry.orderedCapabilities(infrastructureProfileServerBootstrap, infrastructureCapabilityHookBootstrap); err == nil {
		t.Fatal("orderedCapabilities() error = nil, want cycle error")
	}
}

func TestInfrastructureCapabilityRegistryUsesPhaseScopedDependencies(t *testing.T) {
	t.Parallel()

	registry := infrastructureCapabilityRegistry{
		capabilities: map[infrastructureCapabilityName]infrastructureCapability{
			"logger": {
				name:      "logger",
				label:     "logger",
				bootstrap: func(_ context.Context, _ infrastructureProvisioning) error { return nil },
				reload:    func(_ context.Context, _ infrastructureProvisioning) error { return nil },
			},
			"storage": {
				name:               "storage",
				label:              "storage",
				reloadDependencies: []infrastructureCapabilityName{"logger"},
				bootstrap:          func(_ context.Context, _ infrastructureProvisioning) error { return nil },
				reload:             func(_ context.Context, _ infrastructureProvisioning) error { return nil },
			},
		},
		profiles: map[infrastructureCapabilityProfile][]infrastructureCapabilityName{
			infrastructureProfileServerBootstrap: {"storage", "logger"},
			infrastructureProfileRuntimeReload:   {"storage", "logger"},
		},
	}

	bootstrapCapabilities, err := registry.orderedCapabilities(infrastructureProfileServerBootstrap, infrastructureCapabilityHookBootstrap)
	if err != nil {
		t.Fatalf("bootstrap orderedCapabilities() returned error: %v", err)
	}
	if len(bootstrapCapabilities) != 2 {
		t.Fatalf("bootstrap capabilities len = %d, want 2", len(bootstrapCapabilities))
	}
	if bootstrapCapabilities[0].name != "storage" || bootstrapCapabilities[1].name != "logger" {
		t.Fatalf("bootstrap order = [%s %s], want [storage logger]", bootstrapCapabilities[0].name, bootstrapCapabilities[1].name)
	}

	reloadCapabilities, err := registry.orderedCapabilities(infrastructureProfileRuntimeReload, infrastructureCapabilityHookReload)
	if err != nil {
		t.Fatalf("reload orderedCapabilities() returned error: %v", err)
	}
	if len(reloadCapabilities) != 2 {
		t.Fatalf("reload capabilities len = %d, want 2", len(reloadCapabilities))
	}
	if reloadCapabilities[0].name != "logger" || reloadCapabilities[1].name != "storage" {
		t.Fatalf("reload order = [%s %s], want [logger storage]", reloadCapabilities[0].name, reloadCapabilities[1].name)
	}
}
