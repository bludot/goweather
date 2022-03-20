package rediscache

import (
	"context"
	"fmt"
	"github.com/bludot/goweather/config"
	"github.com/bludot/goweather/tracing"
	"log"
	"strconv"
	"time"

	redis "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(config config.RedisDB) *RedisCache {
	return &RedisCache{
		Client: redis.NewClient(&redis.Options{
			Addr:     config.Host + ":" + strconv.Itoa(config.Port),
			Password: config.Password,
			DB:       0,
		}),
	}
}

// for each latitde longitude set range

// ex: 2 	0.01 	1,105.74 	1 km
// so a rang of 10km requires 0.1 difference

// format: long,lat,long,lat where first one is lower and second is higher

func (r RedisCache) SetCache(ctx context.Context, key string, value string) (res bool, err error) {
	method := "SetCache"
	_, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()
	log.Println("got here")
	rdb := r.Client
	set, err := rdb.SetNX(ctx, key, value, 60*60*time.Second).Result()
	if err != nil {
		span.AddSpanError(err)
		span.Log(err.Error())
		span.FailSpan(err.Error())
		return false, err
		// panic(err)
		//return false, err
	}
	return set, nil
}

func (r RedisCache) GetCache(ctx context.Context, key string) (cache string, err error) {
	method := "GetCache"
	_, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()
	rdb := r.Client
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		switch {
		case err == redis.Nil:
			fmt.Println("key does not exist")
			span.AddSpanError(err)
			span.Log(err.Error())
			span.FailSpan(err.Error())
			return
		case err != nil:
			fmt.Println("Get failed", err)
			span.AddSpanError(err)
			span.Log(err.Error())
			span.FailSpan(err.Error())
			return
		case val == "":
			fmt.Println("value is empty")
		}
		// panic(err)
		// return
	}
	return val, nil
}
