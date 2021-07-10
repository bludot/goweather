package main

import (
	"context"
	"fmt"
	"log"
	"time"

	redis "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// for each latitde longitude set range

// ex: 2 	0.01 	1,105.74 	1 km
// so a rang of 10km requires 0.1 difference

// format: long,lat,long,lat where first one is lower and second is higher

func getClient() *redis.Client {
	host := getEnv("REDIS_HOST")
	rdb := redis.NewClient(&redis.Options{
		Addr:     host + "6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return rdb
}

func setCache(key string, value string) (res bool, err error) {
	log.Println("got here")
	rdb := getClient()
	set, err := rdb.SetNX(ctx, key, value, 60*60*time.Second).Result()
	if err != nil {
		panic(err)
		//return false, err
	}
	return set, nil
}

func getCache(key string) (cache string, err error) {
	rdb := getClient()
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		switch {
		case err == redis.Nil:
			fmt.Println("key does not exist")
			return
		case err != nil:
			fmt.Println("Get failed", err)
			return
		case val == "":
			fmt.Println("value is empty")
		}
		// panic(err)
		// return
	}
	fmt.Println(key, val)
	return val, nil
}
