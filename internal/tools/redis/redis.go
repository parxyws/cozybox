package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/parxyws/cozybox/internal/config"
	redisclient "github.com/redis/go-redis/v9"
)

// Client is the global Redis instance
var Client *redisclient.Client

// InitRedis initializes the global Redis client
func InitRedis(cfg *config.Config) error {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	Client = redisclient.NewClient(&redisclient.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,

		// Connection pool best practices
		PoolSize:     100, // Maximum number of socket connections
		MinIdleConns: 10,  // Minimum number of idle connections
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	return nil
}
