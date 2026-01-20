package db

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string) (*redis.Client, error) {
	cl := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	err := cl.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return cl, nil
}
