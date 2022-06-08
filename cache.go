package conveyor

import cmap "github.com/orcaman/concurrent-map"

type Cache struct {
	cmap.ConcurrentMap
}

func NewCache() *Cache {
	return &Cache{
		ConcurrentMap: cmap.New(),
	}
}
