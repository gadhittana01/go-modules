package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Close() error
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
}

type CacheSvc interface {
	Get(ctx context.Context, key string, output any) error
	Set(ctx context.Context, key string, data any, duration ...time.Duration) error
	DelByPrefix(ctx context.Context, prefixName string)
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) time.Duration
	ClearCaches(keys []string, identifier string)
}

type CacheSvcImpl struct {
	config *BaseConfig
	redis  RedisClient
}

func NewCacheSvc(
	config *BaseConfig,
	redis RedisClient,
) CacheSvc {
	return &CacheSvcImpl{
		config: config,
		redis:  redis,
	}
}

func NewRedisClient(config *BaseConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
		Username: config.RedisUsername,
	})

	return rdb
}

func GetOrSetData[T any](c CacheSvc, key string, function func() (T, error), duration ...time.Duration) (T, error) {
	var data T
	ctx := context.Background()
	err := c.Get(ctx, key, &data)
	if err != nil {
		if err == redis.Nil || err.Error() == "Entry not found" {
			data, err := function()
			if err != nil {
				return data, err
			}

			err = c.Set(ctx, key, data, duration...)

			return data, err
		}
		return data, err
	}

	return data, nil
}

func (s *CacheSvcImpl) Get(ctx context.Context, key string, output any) error {
	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		LogInfo(fmt.Sprintf("failed when getting cache (Redis) with key -> %s | error: %v", key, err))
		return err
	}

	err = json.Unmarshal([]byte(val), &output)
	if err != nil {
		LogInfo("failed when unmarshal data")
		return err
	}

	LogInfo(fmt.Sprintf("get data from cache (Redis) with key --> %s", key))

	return nil
}

func (s *CacheSvcImpl) Set(ctx context.Context, key string, data any, duration ...time.Duration) error {
	if data != nil {
		if reflect.TypeOf(data).Kind() == reflect.Slice {
			if reflect.ValueOf(data).Len() == 0 {
				LogInfo("no data to save, array is empty")
				return nil
			}
		}

		cacheData, err := Marshal(data)
		if err != nil {
			return err
		}

		expiration := s.config.CacheDuration
		if len(duration) > 0 {
			expiration = duration[0]
		}

		LogInfo(fmt.Sprintf("set data to cache (Redis) with key --> %s", key))
		return s.redis.Set(ctx, key, cacheData, expiration).Err()
	}

	LogInfo(fmt.Sprintf("not save data to cache, key --> %s", key))

	return nil
}

func (s *CacheSvcImpl) Del(ctx context.Context, key string) error {
	err := s.redis.Del(ctx, key).Err()
	if err != nil {
		LogError(fmt.Sprintf("failed when deleting cache (Redis) with key --> %s", key), zap.Error(err))
		return err
	}

	return nil
}

func (s *CacheSvcImpl) DelByPrefix(ctx context.Context, prefixName string) {
	foundedRecordCount := 0

	iter := s.redis.Scan(ctx, 0, fmt.Sprintf("%s*", prefixName), 0).Iterator()
	LogInfo(fmt.Sprintf("your search pattern: %s", prefixName))

	for iter.Next(ctx) {
		LogInfo(fmt.Sprintf("deleted (Redis)= %s", iter.Val()))
		s.redis.Del(ctx, iter.Val())
		foundedRecordCount++
	}

	if err := iter.Err(); err != nil {
		LogError("failed when deleting cache (Redis)", zap.Error(err))
	}

	LogInfo(fmt.Sprintf("deleted Count (Redis) %d", foundedRecordCount))
}

func (s *CacheSvcImpl) ClearCaches(keys []string, identifier string) {
	ctx := context.Background()
	ewg := errgroup.Group{}
	for i := range keys {
		key := keys[i]
		ewg.Go(func() error {
			if identifier == "" {
				s.DelByPrefix(ctx, BuildPrefixKey(key))
				return nil
			}

			s.DelByPrefix(ctx, BuildPrefixKey(key, identifier))
			return nil
		})
	}

	err := ewg.Wait()
	LogIfError(err)
}

func (s *CacheSvcImpl) Incr(ctx context.Context, key string) (int64, error) {
	res, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return res, err
	}
	return res, nil
}

func (s *CacheSvcImpl) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := s.redis.Expire(ctx, key, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *CacheSvcImpl) TTL(ctx context.Context, key string) time.Duration {
	return s.redis.TTL(ctx, key).Val()
}

func BuildCacheKey(key string, identifier string, funcName string, args ...any) string {
	cacheKey := key
	if identifier != "" && funcName != "" {
		cacheKey = fmt.Sprintf("%s:%s:%s", key, identifier, funcName)
	} else if identifier != "" && funcName == "" {
		cacheKey = fmt.Sprintf("%s:%s", key, identifier)
	} else if identifier == "" && funcName != "" {
		cacheKey = fmt.Sprintf("%s:%s", key, funcName)
	}

	cacheArgs := ""
	for _, arg := range args {

		if cacheArgs != "" {
			cacheArgs += "|"
		}

		v := reflect.ValueOf(arg)
		for i := 0; i < v.NumField(); i++ {
			isEmpty := v.Field(i).Interface() == ""
			if !isEmpty && cacheArgs == "" {
				cacheArgs += fmt.Sprintf("%s->%v", v.Type().Field(i).Name, v.Field(i).Interface())
			} else if !isEmpty && string(cacheArgs[len(cacheArgs)-1]) == "|" {
				cacheArgs += fmt.Sprintf("%s->%v", v.Type().Field(i).Name, v.Field(i).Interface())
			} else if !isEmpty {
				cacheArgs += fmt.Sprintf(",%s->%v", v.Type().Field(i).Name, v.Field(i).Interface())
			}
		}
	}

	return fmt.Sprintf("%s|%s", cacheKey, cacheArgs)
}

func BuildPrefixKey(keys ...string) string {
	prefixKey := ""

	for _, key := range keys {
		if prefixKey == "" {
			prefixKey += key
		} else {
			prefixKey += fmt.Sprintf(":%s", key)
		}
	}

	return prefixKey
}

func InitCacheSvc(t *testing.T, config *BaseConfig) CacheSvc {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return NewCacheSvc(config, redisClient)
}
