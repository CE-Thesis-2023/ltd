package eventsapi

import (
	"labs/local-transcoder/internal/cache"

	"github.com/dgraph-io/ristretto"
	"github.com/panjf2000/ants/v2"
	"labs/local-transcoder/internal/concurrent"
)

type StandardEventHandler struct {
	pool  *ants.Pool
	cache *ristretto.Cache
}

func NewStandardEventHandler() *StandardEventHandler {
	return &StandardEventHandler{
		pool:  custcon.New(100),
		cache: cache.Cache(),
	}
}
