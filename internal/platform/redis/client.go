package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(redisURL string) (*goredis.Client, error) {
	options, err := goredis.ParseURL(redisURL)

	if err != nil {
		return nil, err
	}

	return goredis.NewClient(options), nil
}

func Ping(
	ctx context.Context,
	client *goredis.Client,
) error {
	return client.Ping(ctx).Err()
}
