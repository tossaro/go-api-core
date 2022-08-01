package redis

import (
	"log"
	"time"

	"github.com/go-redis/redis"
)

type (
	Options struct {
		Addrs       []string
		Passwd      string
		Index       int
		PoolSize    *int
		PoolTimeout *time.Duration
		MinIdleConn *int
	}

	Redis struct {
		Cache *redis.Client
	}

	ClusterRedis struct {
		Cache *redis.ClusterClient
	}
)

type Cacher interface {
	Set(k string, p string, v interface{}, d time.Duration) error
	Delete(k string, p string) error
	Get(k string, p string) (v string, err error)
	Ttl(k string, p string) (t time.Duration, err error)
}

const (
	_defaultPoolSize    = 5
	_defaultMinIdleConn = 15
	_defaultPoolTimeout = 5 * time.Second
)

func config(c *Options) *Options {
	if len(c.Addrs) < 1 {
		log.Fatal("redis - address is empty")
	}
	if c.PoolSize == nil {
		poolSize := _defaultPoolSize
		c.PoolSize = &poolSize
	}
	if c.MinIdleConn == nil {
		minIdleConn := _defaultMinIdleConn
		c.MinIdleConn = &minIdleConn
	}
	if c.PoolTimeout != nil {
		poolTimeout := _defaultPoolTimeout
		c.PoolTimeout = &poolTimeout
	}
	return c
}

func NewRedis(o *Options) Cacher {
	o = config(o)
	return &Redis{
		Cache: redis.NewClient(&redis.Options{
			Addr:         o.Addrs[0],
			Password:     o.Passwd,
			DB:           o.Index,
			PoolSize:     *o.PoolSize,
			MinIdleConns: *o.MinIdleConn,
			PoolTimeout:  *o.PoolTimeout,
		}),
	}
}

func NewClusterRedis(o *Options) Cacher {
	o = config(o)
	return &ClusterRedis{
		Cache: redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        o.Addrs,
			Password:     o.Passwd,
			PoolSize:     *o.PoolSize,
			MinIdleConns: *o.MinIdleConn,
			PoolTimeout:  *o.PoolTimeout,
		}),
	}
}

func (r ClusterRedis) Set(k string, p string, v interface{}, d time.Duration) error {
	r.Cache.Del(k + ":" + p)
	if d != 0*time.Second {
		_, err := r.Cache.SetNX(k+":"+p, v, d).Result()
		if err != nil {
			return err
		}
	} else {
		_, err := r.Cache.Set(k+":"+p, v, 0).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r ClusterRedis) Delete(k string, p string) error {
	if cmd := r.Cache.Del(k + ":" + p); cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (r ClusterRedis) Get(k string, p string) (v string, err error) {
	v, err = r.Cache.Get(k + ":" + p).Result()
	if err != nil {
		return v, err
	}
	return v, nil
}

func (r ClusterRedis) Ttl(k string, p string) (t time.Duration, err error) {
	if cmd := r.Cache.TTL(k + ":" + p); cmd.Err() != nil {
		return t, err
	} else {
		return cmd.Val(), nil
	}
}

func (r Redis) Set(k string, p string, v interface{}, d time.Duration) error {
	r.Cache.Del(k + ":" + p)
	if d != 0*time.Second {
		_, err := r.Cache.SetNX(k+":"+p, v, d).Result()
		if err != nil {
			return err
		}
	} else {
		_, err := r.Cache.Set(k+":"+p, v, 0).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r Redis) Delete(k string, p string) error {
	if cmd := r.Cache.Del(k + ":" + p); cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (r Redis) Get(k string, p string) (string, error) {
	v, err := r.Cache.Get(k + ":" + p).Result()
	if err != nil {
		return v, err
	}
	return v, nil
}

func (r Redis) Ttl(k string, p string) (t time.Duration, err error) {
	cmd := r.Cache.TTL(k + ":" + p)
	if err = cmd.Err(); err != nil {
		return t, err
	} else {
		return cmd.Val(), nil
	}
}
