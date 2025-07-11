package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/vishenosik/gocherry/pkg/config"
)

func init() {
	config.AddStructs(RedisConfigEnv{})
}

type RedisConfigEnv struct {
	Host     string        `env:"REDIS_HOST" default:"localhost" desc:"Redis server host"`
	Port     uint16        `env:"REDIS_PORT" default:"6380" desc:"Redis server port"`
	Timeout  time.Duration `env:"REDIS_TIMEOUT" default:"15s" desc:"Redis requests timeout"`
	User     string        `env:"REDIS_USER" default:"user" desc:"Redis user"`
	Password string        `env:"REDIS_USER_PASSWORD" default:"password" desc:"Redis user's password"`
	DB       int           `env:"REDIS_DB" default:"0" desc:"Redis database connection"`
}

func (RedisConfigEnv) Desc() string {
	return "Redis connection settings"
}

type RedisConfig struct {
	Server      config.Server
	Credentials config.Credentials
	DB          int
}

type RedisCache struct {
	client *redis.Client
}

type RedisOption func(*RedisCache)

func validateRedisConfig(config RedisConfig) error {
	const op = "validateConfig"
	if err := config.Server.Validate(); err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}

func NewRedisCache(opts ...RedisOption) (CacheProvider, error) {
	return newRedisCache(opts...)
}

func newRedisCache(opts ...RedisOption) (*RedisCache, error) {
	var envConf RedisConfigEnv
	if err := config.ReadConfig(&envConf); err != nil {
		return nil, errors.Wrap(err, "setup logger: failed to read config")
	}

	return newRedisCacheConfig(RedisConfig{
		Server: config.Server{
			Port: envConf.Port,
			Host: envConf.Host,
		},
		Credentials: config.Credentials{
			User:     envConf.User,
			Password: envConf.Password,
		},
		DB: envConf.DB,
	}, opts...)
}

func newRedisCacheConfig(config RedisConfig, opts ...RedisOption) (*RedisCache, error) {

	if err := validateRedisConfig(config); err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Server.String(),
		Username: config.Credentials.User,
		Password: config.Credentials.Password,
		DB:       config.DB,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	rc := &RedisCache{client: client}

	for _, opt := range opts {
		opt(rc)
	}

	return rc, nil
}

func (rc *RedisCache) Close(_ context.Context) error {
	return rc.client.Close()
}

func (rc *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}
