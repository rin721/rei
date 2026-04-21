package app

import (
	"context"
	"fmt"
)

type sliceCapabilityName string
type sliceCapabilityProfile string
type sliceCapabilityHook string

const (
	sliceCapabilityHookBootstrap sliceCapabilityHook = "bootstrap"
	sliceCapabilityHookStart     sliceCapabilityHook = "start"
	sliceCapabilityHookReload    sliceCapabilityHook = "reload"
	sliceCapabilityHookShutdown  sliceCapabilityHook = "shutdown"
)

const (
	businessCapabilityProfileRuntime sliceCapabilityProfile = "business-runtime"
	deliveryCapabilityProfileRuntime sliceCapabilityProfile = "delivery-runtime"
)

type sliceCapability struct {
	name                  sliceCapabilityName
	label                 string
	bootstrapDependencies []sliceCapabilityName
	startDependencies     []sliceCapabilityName
	reloadDependencies    []sliceCapabilityName
	shutdownDependencies  []sliceCapabilityName
	bootstrap             func(context.Context) error
	start                 func(context.Context) error
	reload                func(context.Context) error
	shutdown              func(context.Context) error
}

type sliceCapabilityRegistry struct {
	capabilities map[sliceCapabilityName]sliceCapability
	profiles     map[sliceCapabilityProfile][]sliceCapabilityName
}

type runtimeStarter struct {
	name  string
	start func(context.Context) error
}

func newRuntimeStarter(name string, start func(context.Context) error) runtimeStarter {
	return runtimeStarter{
		name:  name,
		start: start,
	}
}

func runRuntimeStarters(ctx context.Context, phase string, starters []runtimeStarter) error {
	for _, starter := range starters {
		if err := starter.start(ctx); err != nil {
			return fmt.Errorf("%s: %s: %w", phase, starter.name, err)
		}
	}
	return nil
}

func (c sliceCapability) supports(hook sliceCapabilityHook) bool {
	switch hook {
	case sliceCapabilityHookBootstrap:
		return c.bootstrap != nil
	case sliceCapabilityHookStart:
		return c.start != nil
	case sliceCapabilityHookReload:
		return c.reload != nil
	case sliceCapabilityHookShutdown:
		return c.shutdown != nil
	default:
		return false
	}
}

func (c sliceCapability) dependenciesFor(hook sliceCapabilityHook) []sliceCapabilityName {
	switch hook {
	case sliceCapabilityHookBootstrap:
		return c.bootstrapDependencies
	case sliceCapabilityHookStart:
		return c.startDependencies
	case sliceCapabilityHookReload:
		return c.reloadDependencies
	case sliceCapabilityHookShutdown:
		return c.shutdownDependencies
	default:
		return nil
	}
}

func (r sliceCapabilityRegistry) orderedCapabilities(profile sliceCapabilityProfile, hook sliceCapabilityHook) ([]sliceCapability, error) {
	names := r.profiles[profile]
	selected := make([]sliceCapabilityName, 0, len(names))
	seen := make(map[sliceCapabilityName]struct{}, len(names))

	for _, name := range names {
		capability, ok := r.capabilities[name]
		if !ok {
			return nil, fmt.Errorf("unknown slice capability %q in profile %q", name, profile)
		}
		if !capability.supports(hook) {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		selected = append(selected, name)
	}

	orderedNames, err := orderLifecycleNames(selected, func(name sliceCapabilityName) []sliceCapabilityName {
		return r.capabilities[name].dependenciesFor(hook)
	}, hook == sliceCapabilityHookShutdown)
	if err != nil {
		return nil, fmt.Errorf("cyclic slice capability dependencies in profile %q for %s", profile, hook)
	}

	ordered := make([]sliceCapability, 0, len(orderedNames))
	for _, name := range orderedNames {
		ordered = append(ordered, r.capabilities[name])
	}
	return ordered, nil
}

func (r sliceCapabilityRegistry) bootstrapSteps(profile sliceCapabilityProfile) ([]bootstrapStep, error) {
	capabilities, err := r.orderedCapabilities(profile, sliceCapabilityHookBootstrap)
	if err != nil {
		return nil, err
	}

	steps := make([]bootstrapStep, 0, len(capabilities))
	for _, capability := range capabilities {
		current := capability
		steps = append(steps, newBootstrapStep(current.label, current.bootstrap))
	}
	return steps, nil
}

func (r sliceCapabilityRegistry) starters(profile sliceCapabilityProfile) ([]runtimeStarter, error) {
	capabilities, err := r.orderedCapabilities(profile, sliceCapabilityHookStart)
	if err != nil {
		return nil, err
	}

	starters := make([]runtimeStarter, 0, len(capabilities))
	for _, capability := range capabilities {
		current := capability
		starters = append(starters, newRuntimeStarter(current.label, current.start))
	}
	return starters, nil
}

func (r sliceCapabilityRegistry) reloaders(profile sliceCapabilityProfile) ([]runtimeReloader, error) {
	capabilities, err := r.orderedCapabilities(profile, sliceCapabilityHookReload)
	if err != nil {
		return nil, err
	}

	reloaders := make([]runtimeReloader, 0, len(capabilities))
	for _, capability := range capabilities {
		current := capability
		reloaders = append(reloaders, newRuntimeReloader(current.label, current.reload))
	}
	return reloaders, nil
}

func (r sliceCapabilityRegistry) shutdownSteps(profile sliceCapabilityProfile) ([]shutdownStep, error) {
	capabilities, err := r.orderedCapabilities(profile, sliceCapabilityHookShutdown)
	if err != nil {
		return nil, err
	}

	steps := make([]shutdownStep, 0, len(capabilities))
	for _, capability := range capabilities {
		current := capability
		steps = append(steps, newShutdownStep(current.label, current.shutdown))
	}
	return steps, nil
}

func (p businessProvisioning) lifecycle() sliceCapabilityRegistry {
	return sliceCapabilityRegistry{
		capabilities: map[sliceCapabilityName]sliceCapability{
			"business-modules": {
				name:      "business-modules",
				label:     "business modules",
				bootstrap: func(_ context.Context) error { return p.initRuntime() },
			},
		},
		profiles: map[sliceCapabilityProfile][]sliceCapabilityName{
			businessCapabilityProfileRuntime: {"business-modules"},
		},
	}
}

func (p deliveryProvisioning) lifecycle() sliceCapabilityRegistry {
	return sliceCapabilityRegistry{
		capabilities: map[sliceCapabilityName]sliceCapability{
			"router": {
				name:      "router",
				label:     "router",
				bootstrap: func(_ context.Context) error { return p.initRouter() },
			},
			"http-server": {
				name:                  "http-server",
				label:                 "http server",
				bootstrapDependencies: []sliceCapabilityName{"router"},
				bootstrap:             func(_ context.Context) error { return p.initHTTPServer() },
				start:                 p.startHTTPServer,
				reload:                func(_ context.Context) error { return p.reloadHTTPServer() },
				shutdown:              p.shutdownHTTPServer,
			},
		},
		profiles: map[sliceCapabilityProfile][]sliceCapabilityName{
			deliveryCapabilityProfileRuntime: {"router", "http-server"},
		},
	}
}
