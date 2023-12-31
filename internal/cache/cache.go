package cache

import (
	"log"
	"sync"

	"github.com/dgraph-io/ristretto"
)

var once sync.Once

var cache *ristretto.Cache

func Init() {
	once.Do(func() {
		store, err := ristretto.NewCache(&ristretto.Config{
			NumCounters: 5000000,
			MaxCost:     5000000,
			BufferItems: 64,
		})
		if err != nil {
			log.Fatalf("cache.Init: err = %s", err)
			return
		}

		cache = store
	})
}

func Stop() {
	cache.Close()
}

func Cache() *ristretto.Cache {
	return cache
}
