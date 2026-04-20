package app

import pkgcache "github.com/rin721/rei/pkg/cache"

func (p infrastructureProvisioning) initCache() error {
	if p.infra.cache != nil {
		return nil
	}

	p.infra.cache = pkgcache.New(toCacheConfig(p.cfg.Redis))
	return nil
}
