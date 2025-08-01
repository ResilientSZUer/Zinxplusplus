package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"zinxplusplus/ziface"

	"github.com/go-redis/redis/v8"
)

type RedisStateAdapter struct {
	client *redis.Client
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

func NewRedisStateAdapter(cfg RedisConfig) (ziface.IStateManager, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis (%s): %w", cfg.Addr, err)
	}

	fmt.Printf("[State] RedisStateAdapter connected to %s, DB %d\n", cfg.Addr, cfg.DB)

	return &RedisStateAdapter{
		client: rdb,
	}, nil
}

func (rsa *RedisStateAdapter) SetState(ctx context.Context, key string, value []byte, expiration int64) error {
	var expDuration time.Duration
	if expiration > 0 {

		expDuration = time.Duration(expiration) * time.Second
	} else {
		expDuration = 0
	}

	err := rsa.client.Set(ctx, key, value, expDuration).Err()
	if err != nil {

		return fmt.Errorf("%w: set key %s: %v", ErrRedisCmdFailed, key, err)
	}
	return nil
}

func (rsa *RedisStateAdapter) GetState(ctx context.Context, key string) ([]byte, error) {
	val, err := rsa.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {

			return nil, fmt.Errorf("%w: key=%s", ErrStateNotFound, key)
		}

		return nil, fmt.Errorf("%w: get key %s: %v", ErrRedisCmdFailed, key, err)
	}
	return val, nil
}

func (rsa *RedisStateAdapter) DeleteState(ctx context.Context, key string) error {
	err := rsa.client.Del(ctx, key).Err()
	if err != nil {

		return fmt.Errorf("%w: del key %s: %v", ErrRedisCmdFailed, key, err)
	}
	return nil
}

func (rsa *RedisStateAdapter) ExistsState(ctx context.Context, key string) (bool, error) {
	val, err := rsa.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("%w: exists key %s: %v", ErrRedisCmdFailed, key, err)
	}
	return val > 0, nil
}

func (rsa *RedisStateAdapter) SetStateObject(ctx context.Context, key string, obj interface{}, expiration int64) error {

	valueBytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("%w: key=%s, type=%T: %v", ErrSerializationFailed, key, obj, err)
	}

	return rsa.SetState(ctx, key, valueBytes, expiration)
}

func (rsa *RedisStateAdapter) GetStateObject(ctx context.Context, key string, objPtr interface{}) error {

	valueBytes, err := rsa.GetState(ctx, key)
	if err != nil {

		return err
	}

	if err := json.Unmarshal(valueBytes, objPtr); err != nil {
		return fmt.Errorf("%w: key=%s, targetType=%T: %v", ErrDeserializationFailed, key, objPtr, err)
	}

	return nil
}
