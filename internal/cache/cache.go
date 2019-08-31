package cache

import (
	"time"

	rediscache "github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

// nolint: gochecknoglobals
var cacheGets = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "cache_get_total",
		Help: "Total number of cache gets",
	},
)

// nolint: gochecknoglobals
var cachePuts = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "cache_put_total",
		Help: "Total number of cache puts",
	},
)

// nolint: gochecknoinits
func init() {
	prometheus.MustRegister(cacheGets, cachePuts)
}

// Redis cache
type Redis struct {
	redis *redis.Client
	codec *rediscache.Codec
}

// New redis cache
func New(redis *redis.Client) *Redis {
	codec := &rediscache.Codec{
		Redis: redis,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	return &Redis{
		redis: redis,
		codec: codec,
	}
}

// Close connections
func (c *Redis) Close() error {
	return c.redis.Close()
}

// Get from cache by key
func (c *Redis) Get(key string, result interface{}) (err error) {
	cacheGets.Inc()
	return c.codec.Get(key, result)
}

// Put on cache
func (c *Redis) Put(key string, obj interface{}, expire time.Duration) (err error) {
	cachePuts.Inc()
	return c.codec.Set(&rediscache.Item{
		Key:        key,
		Object:     obj,
		Expiration: expire,
	})
}
