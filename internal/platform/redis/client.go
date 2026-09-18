package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(redisURL string) (*goredis.Client, error) {
	options, err := goredis.ParseURL(redisURL)

	if err != nil {
		return nil, fmt.Errorf("interpretar URL de Redis: %w", err)
	}

	return goredis.NewClient(options), nil
}

func Ping(
	ctx context.Context,
	client *goredis.Client,
) error {
	// Se reutiliza para la disponibilidad inicial y la ruta de readiness.
	return client.Ping(ctx).Err()
}
