package cache

import (
	rediscache "github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

// nolint: gochecknoglobals
// 缓存获取gets_total
var cacheGets = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "cache", // 子系统
		Name:      "gets_total",
		Help:      "Total number of successful cache gets",
	},
)

// nolint: gochecknoglobals
// 缓存puts_total
var cachePuts = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "cache",
		Name:      "puts_total",
		Help:      "Total number of successful cache puts",
	},
)

// nolint: gochecknoglobals
// 缓存删除deletes_total
var cacheDeletes = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "cache",
		Name:      "deletes_total",
		Help:      "Total number of successful cache deletes",
	},
)

// nolint: gochecknoinits
// 初始化函数
func init() {
	// MustRegister用 DefaultRegisterer 对提供的收集器进行注册
	prometheus.MustRegister(cacheGets, cachePuts, cacheDeletes)
}

// Redis cache.
// redis缓存实例（包含客户端，redis缓存）
type Redis struct {
	redis *redis.Client     // redis客户端配置
	codec *rediscache.Codec // redis缓存配置
}

// New redis cache.
// 新建redis缓存
func New(redis *redis.Client) *Redis {
	codec := &rediscache.Codec{
		Redis: redis,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		}, // 序列化
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		}, // 反序列化
	}

	return &Redis{
		redis: redis, // 客户端配置赋值
		codec: codec, // redis缓存配置
	}
}

// Close connections.
// func (c *baseClient) Close() error
func (c *Redis) Close() error {
	return c.redis.Close()
}

// Get from cache by key.
func (c *Redis) Get(key string, result interface{}) error {
	if err := c.codec.Get(key, result); err != nil {
		return err
	}
	// atomic.AddUint64(&c.valInt, 1)
	cacheGets.Inc() // 缓存查询新增
	return nil
}

// Put on cache.
func (c *Redis) Put(key string, obj interface{}) error {
	if err := c.codec.Set(&rediscache.Item{
		Key:    key,
		Object: obj,
	}); err != nil {
		return err
	}
	cachePuts.Inc()
	return nil
}

// Delete from cache.
func (c *Redis) Delete(key string) error {
	if err := c.codec.Delete(key); err != nil {
		return err
	}
	cacheDeletes.Inc()
	return nil
}
