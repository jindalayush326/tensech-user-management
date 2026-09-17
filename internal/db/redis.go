package db

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(addr string, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Println("redis connection failed:", err)
		return client
	}

	log.Println("connected to redis")

	return client
}