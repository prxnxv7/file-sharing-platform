package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
    redisClient *redis.Client
    ctx = context.Background()
)

func InitRedis() {
    redisClient = redis.NewClient(&redis.Options{
        Addr:     "redis:6379",
        Password: "",
        DB:       0, // Default DB
    })

    	// Test connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	} else {
		fmt.Println("Connected to Redis successfully!")
	}
}

func CacheFileMetadata(key string, data string, expiration time.Duration) error {
    err := redisClient.Set(ctx, key, data, expiration).Err()
    if err != nil {
        return fmt.Errorf("failed to cache file metadata: %v", err)
    }
    fmt.Printf("File metadata for key %s cached successfully!\n", key)
    return nil
}

func GetCachedFileMetadata(key string) (string, error) {
    val, err := redisClient.Get(ctx, key).Result()
    if err == redis.Nil {
        fmt.Printf("No cached data for key %s\n", key)
        return "", nil
    }
    if err != nil {
        return "", fmt.Errorf("failed to retrieve cached file metadata: %v", err)
    }
    fmt.Printf("Cached file metadata for key %s retrieved successfully!\n", key)
    return val, nil
}

func InvalidateCache(key string) error {
    err := redisClient.Del(ctx, key).Err()
    if err != nil {
        return fmt.Errorf("failed to invalidate cache for key %s: %v", key, err)
    }
    fmt.Printf("Cache for key %s invalidated successfully!\n", key)
    return nil
}

// Rate limiting for API calls
func RateLimit(userID string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s", userID)
	count, err := redisClient.Get(ctx, key).Int()
	if err == redis.Nil {
		err = redisClient.Set(ctx, key, 1, window).Err()
		if err != nil {
			return false, err
		}
		return true, nil
	} else if err != nil {
		return false, err
	}

	if count >= limit {
		return false, nil
	}

	redisClient.Incr(ctx, key)
	return true, nil
}
