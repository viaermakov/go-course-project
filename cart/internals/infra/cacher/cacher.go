package cacher

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type Cacher struct {
	client *redis.Client
	ttl    time.Duration
	waitCh chan struct{}
}

func New(ctx context.Context, addr string, ttl time.Duration) *Cacher {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	waitCh := make(chan struct{}, 1)
	waitCh <- struct{}{}

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(waitCh)
				log.Println("[WARNING] cacher is shutting down")

				if err := client.Close(); err != nil {
					log.Println("[ERROR] close redis client:", err)
				}
			default:
			}
		}
	}()

	return &Cacher{
		client: client,
		ttl:    ttl,
		waitCh: waitCh,
	}
}

func (c *Cacher) Add(ctx context.Context, key string, value []byte) error {
	return c.client.SetNX(ctx, key, value, c.ttl).Err()
}

func (c *Cacher) Get(ctx context.Context, key string) ([]byte, error) {
	v := c.client.Get(ctx, key)
	err := v.Err()

	if err == redis.Nil {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	b, err := v.Bytes()

	return b, err
}

func (c *Cacher) Lock() {
	<-c.waitCh
}

func (c *Cacher) Unlock() {
	c.waitCh <- struct{}{}
}
