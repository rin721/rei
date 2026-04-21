package app

import (
	"context"
	"testing"
)

func TestSliceCapabilityRegistryBootstrapOrdersDependencies(t *testing.T) {
	t.Parallel()

	registry := sliceCapabilityRegistry{
		capabilities: map[sliceCapabilityName]sliceCapability{
			"router": {
				name:      "router",
				label:     "router",
				bootstrap: func(_ context.Context) error { return nil },
			},
			"http-server": {
				name:                  "http-server",
				label:                 "http server",
				bootstrapDependencies: []sliceCapabilityName{"router"},
				bootstrap:             func(_ context.Context) error { return nil },
			},
		},
		profiles: map[sliceCapabilityProfile][]sliceCapabilityName{
			deliveryCapabilityProfileRuntime: {"http-server", "router"},
		},
	}

	steps, err := registry.bootstrapSteps(deliveryCapabilityProfileRuntime)
	if err != nil {
		t.Fatalf("bootstrapSteps() returned error: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("bootstrap steps len = %d, want 2", len(steps))
	}
	if steps[0].name != "router" || steps[1].name != "http server" {
		t.Fatalf("bootstrap order = [%s %s], want [router http server]", steps[0].name, steps[1].name)
	}
}

func TestSliceCapabilityRegistryStartUsesOnlyStartHooks(t *testing.T) {
	t.Parallel()

	registry := sliceCapabilityRegistry{
		capabilities: map[sliceCapabilityName]sliceCapability{
			"router": {
				name:      "router",
				label:     "router",
				bootstrap: func(_ context.Context) error { return nil },
			},
			"http-server": {
				name:  "http-server",
				label: "http server",
				start: func(_ context.Context) error { return nil },
			},
		},
		profiles: map[sliceCapabilityProfile][]sliceCapabilityName{
			deliveryCapabilityProfileRuntime: {"router", "http-server"},
		},
	}

	starters, err := registry.starters(deliveryCapabilityProfileRuntime)
	if err != nil {
		t.Fatalf("starters() returned error: %v", err)
	}
	if len(starters) != 1 {
		t.Fatalf("starters len = %d, want 1", len(starters))
	}
	if starters[0].name != "http server" {
		t.Fatalf("starter name = %q, want %q", starters[0].name, "http server")
	}
}
