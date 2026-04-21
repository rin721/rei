package app

import (
	"context"
	"fmt"
)

type infrastructureCapabilityName string
type infrastructureCapabilityProfile string

const (
	infrastructureCapabilityLogger      infrastructureCapabilityName = "logger"
	infrastructureCapabilityI18n        infrastructureCapabilityName = "i18n"
	infrastructureCapabilityIDGenerator infrastructureCapabilityName = "id-generator"
	infrastructureCapabilityCache       infrastructureCapabilityName = "cache"
	infrastructureCapabilityDatabase    infrastructureCapabilityName = "database"
	infrastructureCapabilityDBTx        infrastructureCapabilityName = "dbtx"
	infrastructureCapabilityExecutor    infrastructureCapabilityName = "executor"
	infrastructureCapabilityCrypto      infrastructureCapabilityName = "crypto"
	infrastructureCapabilityJWT         infrastructureCapabilityName = "jwt"
	infrastructureCapabilityStorage     infrastructureCapabilityName = "storage"
	infrastructureCapabilityRBAC        infrastructureCapabilityName = "rbac"
	infrastructureCapabilityConfig      infrastructureCapabilityName = "config-manager"
)

const (
	infrastructureProfileServerBootstrap infrastructureCapabilityProfile = "server-bootstrap"
	infrastructureProfileDBBootstrap     infrastructureCapabilityProfile = "db-bootstrap"
	infrastructureProfileRuntimeReload   infrastructureCapabilityProfile = "runtime-reload"
	infrastructureProfileRuntimeShutdown infrastructureCapabilityProfile = "runtime-shutdown"
)

type infrastructureCapability struct {
	name                  infrastructureCapabilityName
	label                 string
	bootstrapDependencies []infrastructureCapabilityName
	reloadDependencies    []infrastructureCapabilityName
	shutdownDependencies  []infrastructureCapabilityName
	bootstrap             func(context.Context, infrastructureProvisioning) error
	reload                func(context.Context, infrastructureProvisioning) error
	shutdown              func(context.Context, infrastructureProvisioning) error
}

type infrastructureCapabilityRegistry struct {
	capabilities map[infrastructureCapabilityName]infrastructureCapability
	profiles     map[infrastructureCapabilityProfile][]infrastructureCapabilityName
}

type infrastructureCapabilityHook string

const (
	infrastructureCapabilityHookBootstrap infrastructureCapabilityHook = "bootstrap"
	infrastructureCapabilityHookReload    infrastructureCapabilityHook = "reload"
	infrastructureCapabilityHookShutdown  infrastructureCapabilityHook = "shutdown"
)

func newInfrastructureCapabilityRegistry() infrastructureCapabilityRegistry {
	capabilities := map[infrastructureCapabilityName]infrastructureCapability{
		infrastructureCapabilityLogger: {
			name:  infrastructureCapabilityLogger,
			label: "logger",
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initLogger()
			},
			reload: func(_ context.Context, p infrastructureProvisioning) error {
				return p.reloadLogger()
			},
			shutdown: func(_ context.Context, p infrastructureProvisioning) error {
				return p.flushLogger()
			},
		},
		infrastructureCapabilityI18n: {
			name:  infrastructureCapabilityI18n,
			label: "i18n",
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initI18n()
			},
		},
		infrastructureCapabilityIDGenerator: {
			name:  infrastructureCapabilityIDGenerator,
			label: "id generator",
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initIDGenerator()
			},
		},
		infrastructureCapabilityCache: {
			name:  infrastructureCapabilityCache,
			label: "cache",
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initCache()
			},
			reload: func(_ context.Context, p infrastructureProvisioning) error {
				return p.reloadCache()
			},
			shutdown: func(_ context.Context, p infrastructureProvisioning) error {
				return p.closeCache()
			},
		},
		infrastructureCapabilityDatabase: {
			name:  infrastructureCapabilityDatabase,
			label: "database",
			bootstrap: func(ctx context.Context, p infrastructureProvisioning) error {
				return p.initDatabase(ctx)
			},
			reload: func(_ context.Context, p infrastructureProvisioning) error {
				return p.reloadDatabase()
			},
			shutdown: func(_ context.Context, p infrastructureProvisioning) error {
				return p.closeDatabase()
			},
		},
		infrastructureCapabilityDBTx: {
			name:                  infrastructureCapabilityDBTx,
			label:                 "database transaction manager",
			bootstrapDependencies: []infrastructureCapabilityName{infrastructureCapabilityDatabase},
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initDBTx()
			},
		},
		infrastructureCapabilityExecutor: {
			name:  infrastructureCapabilityExecutor,
			label: "executor",
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initExecutor()
			},
			reload: func(_ context.Context, p infrastructureProvisioning) error {
				return p.reloadExecutor()
			},
			shutdown: func(ctx context.Context, p infrastructureProvisioning) error {
				return p.shutdownExecutor(ctx)
			},
		},
		infrastructureCapabilityCrypto: {
			name:  infrastructureCapabilityCrypto,
			label: "crypto",
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initCrypto()
			},
		},
		infrastructureCapabilityJWT: {
			name:  infrastructureCapabilityJWT,
			label: "jwt",
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initJWT()
			},
		},
		infrastructureCapabilityStorage: {
			name:  infrastructureCapabilityStorage,
			label: "storage",
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initStorage()
			},
			reload: func(ctx context.Context, p infrastructureProvisioning) error {
				return p.reloadStorage(ctx)
			},
			shutdown: func(_ context.Context, p infrastructureProvisioning) error {
				return p.closeStorage()
			},
		},
		infrastructureCapabilityRBAC: {
			name:  infrastructureCapabilityRBAC,
			label: "rbac",
			bootstrap: func(_ context.Context, p infrastructureProvisioning) error {
				return p.initRBAC()
			},
			shutdown: func(_ context.Context, p infrastructureProvisioning) error {
				return p.closeRBAC()
			},
		},
		infrastructureCapabilityConfig: {
			name:  infrastructureCapabilityConfig,
			label: "config manager",
			shutdown: func(_ context.Context, p infrastructureProvisioning) error {
				return p.stopConfigManager()
			},
		},
	}

	return infrastructureCapabilityRegistry{
		capabilities: capabilities,
		profiles: map[infrastructureCapabilityProfile][]infrastructureCapabilityName{
			infrastructureProfileServerBootstrap: {
				infrastructureCapabilityLogger,
				infrastructureCapabilityI18n,
				infrastructureCapabilityIDGenerator,
				infrastructureCapabilityCache,
				infrastructureCapabilityDatabase,
				infrastructureCapabilityDBTx,
				infrastructureCapabilityExecutor,
				infrastructureCapabilityCrypto,
				infrastructureCapabilityJWT,
				infrastructureCapabilityStorage,
				infrastructureCapabilityRBAC,
			},
			infrastructureProfileDBBootstrap: {
				infrastructureCapabilityLogger,
				infrastructureCapabilityDatabase,
			},
			infrastructureProfileRuntimeReload: {
				infrastructureCapabilityLogger,
				infrastructureCapabilityCache,
				infrastructureCapabilityDatabase,
				infrastructureCapabilityExecutor,
				infrastructureCapabilityStorage,
			},
			infrastructureProfileRuntimeShutdown: {
				infrastructureCapabilityConfig,
				infrastructureCapabilityStorage,
				infrastructureCapabilityExecutor,
				infrastructureCapabilityCache,
				infrastructureCapabilityDatabase,
				infrastructureCapabilityRBAC,
				infrastructureCapabilityLogger,
			},
		},
	}
}

func (p infrastructureProvisioning) capabilities() infrastructureCapabilityRegistry {
	return newInfrastructureCapabilityRegistry()
}

func (c infrastructureCapability) supports(hook infrastructureCapabilityHook) bool {
	switch hook {
	case infrastructureCapabilityHookBootstrap:
		return c.bootstrap != nil
	case infrastructureCapabilityHookReload:
		return c.reload != nil
	case infrastructureCapabilityHookShutdown:
		return c.shutdown != nil
	default:
		return false
	}
}

func (c infrastructureCapability) dependenciesFor(hook infrastructureCapabilityHook) []infrastructureCapabilityName {
	switch hook {
	case infrastructureCapabilityHookBootstrap:
		return c.bootstrapDependencies
	case infrastructureCapabilityHookReload:
		return c.reloadDependencies
	case infrastructureCapabilityHookShutdown:
		return c.shutdownDependencies
	default:
		return nil
	}
}

func (r infrastructureCapabilityRegistry) orderedCapabilities(profile infrastructureCapabilityProfile, hook infrastructureCapabilityHook) ([]infrastructureCapability, error) {
	names := r.profiles[profile]
	selected := make([]infrastructureCapabilityName, 0, len(names))

	for _, name := range names {
		capability, ok := r.capabilities[name]
		if !ok {
			return nil, fmt.Errorf("unknown infrastructure capability %q in profile %q", name, profile)
		}
		if !capability.supports(hook) {
			continue
		}
		duplicate := false
		for _, selectedName := range selected {
			if selectedName == name {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		selected = append(selected, name)
	}

	orderedNames, err := orderLifecycleNames(selected, func(name infrastructureCapabilityName) []infrastructureCapabilityName {
		return r.capabilities[name].dependenciesFor(hook)
	}, hook == infrastructureCapabilityHookShutdown)
	if err != nil {
		return nil, fmt.Errorf("cyclic infrastructure capability dependencies in profile %q for %s", profile, hook)
	}

	ordered := make([]infrastructureCapability, 0, len(orderedNames))
	for _, name := range orderedNames {
		ordered = append(ordered, r.capabilities[name])
	}
	return ordered, nil
}

func (r infrastructureCapabilityRegistry) bootstrapSteps(profile infrastructureCapabilityProfile, p infrastructureProvisioning) ([]bootstrapStep, error) {
	capabilities, err := r.orderedCapabilities(profile, infrastructureCapabilityHookBootstrap)
	if err != nil {
		return nil, err
	}

	steps := make([]bootstrapStep, 0, len(capabilities))
	for _, capability := range capabilities {
		current := capability
		steps = append(steps, newBootstrapStep(current.label, func(ctx context.Context) error {
			return current.bootstrap(ctx, p)
		}))
	}
	return steps, nil
}

func (r infrastructureCapabilityRegistry) reloaders(profile infrastructureCapabilityProfile, p infrastructureProvisioning) ([]runtimeReloader, error) {
	capabilities, err := r.orderedCapabilities(profile, infrastructureCapabilityHookReload)
	if err != nil {
		return nil, err
	}

	reloaders := make([]runtimeReloader, 0, len(capabilities))
	for _, capability := range capabilities {
		current := capability
		reloaders = append(reloaders, newRuntimeReloader(current.label, func(ctx context.Context) error {
			return current.reload(ctx, p)
		}))
	}
	return reloaders, nil
}

func (r infrastructureCapabilityRegistry) shutdownSteps(profile infrastructureCapabilityProfile, p infrastructureProvisioning) ([]shutdownStep, error) {
	capabilities, err := r.orderedCapabilities(profile, infrastructureCapabilityHookShutdown)
	if err != nil {
		return nil, err
	}

	steps := make([]shutdownStep, 0, len(capabilities))
	for _, capability := range capabilities {
		current := capability
		steps = append(steps, newShutdownStep(current.label, func(ctx context.Context) error {
			return current.shutdown(ctx, p)
		}))
	}
	return steps, nil
}
