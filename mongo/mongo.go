package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Options struct {
	Host        string
	Port        string
	MaxPoolSize *uint64
	MinPoolSize *uint64
	Timeout     *time.Duration
	MaxIdleTime *time.Duration
	Schema      string
}

const (
	_defaultPoolMax     = uint64(3)
	_defaultPoolMin     = uint64(1)
	_defaultTimeout     = 2 * time.Second
	_defaultMaxConnIdle = 10 * time.Second
)

func config(c *Options) *Options {
	if c.MaxPoolSize == nil {
		poolSize := _defaultPoolMax
		c.MaxPoolSize = &poolSize
	}
	if c.MinPoolSize == nil {
		poolSize := _defaultPoolMin
		c.MinPoolSize = &poolSize
	}
	if c.Timeout != nil {
		timeout := _defaultTimeout
		c.Timeout = &timeout
	}
	if c.MaxIdleTime != nil {
		connIdle := _defaultMaxConnIdle
		c.MaxIdleTime = &connIdle
	}
	return c
}

func New(o *Options) (*mongo.Database, error) {
	o = config(o)
	opts := options.Client()
	opts.SetMaxPoolSize(*o.MaxPoolSize)
	opts.SetMinPoolSize(*o.MinPoolSize)
	opts.SetConnectTimeout(*o.Timeout)
	opts.SetMaxConnIdleTime(*o.MaxIdleTime)
	opts.ApplyURI("mongodb://" + o.Host + ":" + o.Port)
	client, err := mongo.NewClient(opts)
	if err != nil {
		log.Fatal("mongo - failed to settle client %w", err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal("mongo - failed to settle connection %w", err)
	}
	return client.Database(o.Schema), nil
}
