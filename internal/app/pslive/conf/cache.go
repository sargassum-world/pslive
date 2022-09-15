package conf

import (
	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/env"
)

const cacheEnvPrefix = "CACHE_"

func getCacheConfig() (c ristretto.Config, err error) {
	const defaultNumCounters = 3e6 // default: 300k items, ~9 MB of counters
	c.NumCounters, err = env.GetInt64(cacheEnvPrefix+"CACHE_NUMCOUNTERS", defaultNumCounters)
	if err != nil {
		return ristretto.Config{}, errors.Wrap(err, "couldn't make numCounters config")
	}

	const defaultMaxCost = 3e7 // default: up to 30 MB total with min cost weight of 1
	c.MaxCost, err = env.GetInt64(cacheEnvPrefix+"MAXCOST", defaultMaxCost)
	if err != nil {
		return ristretto.Config{}, errors.Wrap(err, "couldn't make maxCost config")
	}

	const defaultBufferItems = 64 // default: ristretto's recommended value
	c.BufferItems, err = env.GetInt64(cacheEnvPrefix+"BUFFERITEMS", defaultBufferItems)
	if err != nil {
		return ristretto.Config{}, errors.Wrap(err, "couldn't make bufferItems config")
	}

	c.Metrics, err = env.GetBool(cacheEnvPrefix + "METRICS")
	if err != nil {
		return ristretto.Config{}, errors.Wrap(err, "couldn't make metrics config")
	}
	return c, nil
}
