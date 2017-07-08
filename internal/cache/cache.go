package cache

import (
	"time"

	"github.com/apex/log"
	rediscache "github.com/go-redis/cache"
	"github.com/go-redis/redis"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

// Redis cache
type Redis struct {
	redis *redis.Client
	codec *rediscache.Codec
}

// New redis cache
func New(url string) *Redis {
	options, err := redis.ParseURL(url)
	if err != nil {
		log.WithError(err).Fatal("invalid redis_url")
	}
	var redis = redis.NewClient(options)
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
	return c.codec.Get(key, result)
}

// Put on cache
func (c *Redis) Put(key string, obj interface{}, expire time.Duration) (err error) {
	return c.codec.Set(&rediscache.Item{
		Key:        key,
		Object:     obj,
		Expiration: expire,
	})
}
