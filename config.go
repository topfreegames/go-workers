package workers

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Options struct {
	Address      string
	Password     string
	Database     string
	ProcessID    string
	Namespace    string
	PoolSize     int
	PoolInterval int
	DialOptions  []redis.DialOption
}

type WorkerConfig struct {
	processId    string
	Namespace    string
	PoolInterval int
	Pool         *redis.Pool
	Fetch        func(queue string) Fetcher
}

var Config *WorkerConfig

func Configure(options Options) {
	var namespace string

	if options.Address == "" {
		panic("Configure requires a 'Address' option, which identifies a Redis instance")
	}
	if options.ProcessID == "" {
		panic("Configure requires a 'ProcessID' option, which uniquely identifies this instance")
	}
	if options.PoolSize <= 0 {
		options.PoolSize = 1
	}
	if options.Namespace != "" {
		namespace = options.Namespace + ":"
	}
	if options.PoolInterval <= 0 {
		options.PoolInterval = 15
	}

	Config = &WorkerConfig{
		options.ProcessID,
		namespace,
		options.PoolInterval,
		GetConnectionPool(options),
		func(queue string) Fetcher {
			return NewFetch(queue, make(chan *Msg), make(chan bool))
		},
	}
}

func GetConnectionPool(options Options) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     options.PoolSize,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", options.Address, options.DialOptions...)
			if err != nil {
				return nil, err
			}
			if options.Password != "" {
				if _, err := c.Do("AUTH", options.Password); err != nil {
					if errClose := c.Close(); errClose != nil {
						return nil, fmt.Errorf("%w. failed to close connection: %s", err, errClose.Error())
					}
					return nil, err
				}
			}
			if options.Database != "" {
				if _, err := c.Do("SELECT", options.Database); err != nil {
					if errClose := c.Close(); errClose != nil {
						return nil, fmt.Errorf("%w. failed to close connection: %s", err, errClose.Error())
					}
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
