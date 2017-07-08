package cache

import (
	"time"

	rediscache "github.com/go-redis/cache"
	"github.com/go-redis/redis"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

// Redis cache
type Redis struct {
	ring  *redis.Ring
	codec *rediscache.Codec
}

// New redis cache
func New(url string) *Redis {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server": url,
		},
	})
	codec := &rediscache.Codec{
		Redis: ring,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	return &Redis{
		ring:  ring,
		codec: codec,
	}
}

// Close connections
func (c *Redis) Close() error {
	return c.ring.Close()
}

// Get from cache by key
func (c *Redis) Get(key string, result interface{}) (err error) {
	return c.codec.Get(key, result)
}

// Put on cache
func (c *Redis) Put(key string, obj interface{}) (err error) {
	return c.codec.Set(&rediscache.Item{
		Key:        key,
		Object:     obj,
		Expiration: time.Hour * 2,
	})
}
