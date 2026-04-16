package app

import pkgcache "github.com/rin721/rei/pkg/cache"

func (a *App) initCache() error {
	if a.cache != nil {
		return nil
	}

	a.cache = pkgcache.New(toCacheConfig(a.cfg.Redis))
	return nil
}
