package conveyor

import cmap "github.com/orcaman/concurrent-map"

type Cache struct {
	cmap.ConcurrentMap
}

func newCache() *Cache {
	return &Cache{
		ConcurrentMap: cmap.New(),
	}
}
