package controllers

import (
	"github.com/bludot/goweather/rediscache"
	"github.com/bludot/goweather/weatherapi"
)

type Controller struct {
	RedisCache *rediscache.RedisCache
	WeatherApi *weatherapi.WeatherAPI
}

func NewController(redisCache *rediscache.RedisCache, weatherapi *weatherapi.WeatherAPI) *Controller {
	return &Controller{
		RedisCache: redisCache,
		WeatherApi: weatherapi,
	}
}
