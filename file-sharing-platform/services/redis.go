package services

import (
    "context"
    "github.com/go-redis/redis/v8"
    "os"
    "time"
)

var (
    redisClient *redis.Client
    ctx = context.Background()
)

func init() {
    redisClient = redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_ADDR"),
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0, // Default DB
    })
}

func CacheFileMetadata(key string, data string, expiration time.Duration) error {
    err := redisClient.Set(ctx, key, data, expiration).Err()
    if err != nil {
        return err
    }
    return nil
}

func GetCachedFileMetadata(key string) (string, error) {
    val, err := redisClient.Get(ctx, key).Result()
    if err == redis.Nil {
        return "", nil
    }
    if err != nil {
        return "", err
    }
    return val, nil
}

func InvalidateCache(key string) error {
    err := redisClient.Del(ctx, key).Err()
    if err != nil {
        return err
    }
    return nil
}
