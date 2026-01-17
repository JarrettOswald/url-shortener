package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository interface {
	Save(ctx context.Context, key string, url string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type redisRepo struct {
	client *redis.Client
}

func (rr *redisRepo) Get(ctx context.Context, key string) (string, error) {
	return rr.client.Get(ctx, key).Result()
}

func (rr *redisRepo) Save(ctx context.Context, key string, url string, ttl time.Duration) error {
	return rr.client.Set(ctx, key, url, ttl).Err()
}

func NewRedisRepository(client *redis.Client) Repository {
	return &redisRepo{client: client}
}

type Config struct {
	Addr        string        `yaml:"addr"`
	Password    string        `yaml:"password"`
	User        string        `yaml:"user"`
	DB          int           `yaml:"db"`
	MaxRetries  int           `yaml:"max_retries"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Timeout     time.Duration `yaml:"timeout"`
}

func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		Username:     cfg.User,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	if err := db.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}

	return db, nil
}
