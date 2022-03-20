package weatherapi

import (
	"context"
	"fmt"
	"github.com/bludot/goweather/config"
	"github.com/bludot/goweather/http_client"
	"github.com/bludot/goweather/rediscache"
	"io/ioutil"
	"log"
	"net/http"
)

type Location struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	IP        int     `json:"ip"`
}

type WeatherAPI struct {
	APIKey     string
	RedisCache *rediscache.RedisCache
	HttpClient http_client.HTTPClient
}

func NewWeatherAPI(config config.WeatherAPIConfig, redisCache *rediscache.RedisCache) *WeatherAPI {
	return &WeatherAPI{
		APIKey:     config.APIKey,
		RedisCache: redisCache,
		HttpClient: http_client.NewClient(http.DefaultClient),
	}
}

func (w WeatherAPI) GetCurrentWeather(ctx context.Context, location *Location) (res *string, failed error) {
	key := w.GetCity(ctx, location).City + "_current"
	// key := fmt.Sprintf("current%f,%f", location.Longitude, location.Latitude)
	cache, err := w.RedisCache.GetCache(key)
	if err != nil {
		log.Println("got here")
		// return ""
	}
	if err == nil {
		return &cache, nil
	}
	apikey := w.APIKey
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%f,%f", apikey, location.Latitude, location.Longitude)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := w.HttpClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	w.RedisCache.SetCache(key, sb)
	// log.Printf(sb)
	return &sb, nil
}

func (w WeatherAPI) GetForecast(ctx context.Context, location *Location) (res *string, failed error) {
	key := w.GetCity(ctx, location).City + "_forecast"
	// key := fmt.Sprintf("forecast%f,%f", location.Longitude, location.Latitude)
	cache, err := w.RedisCache.GetCache(key)
	if err != nil {
		log.Println("got here")
		// return ""
	}
	if err == nil {
		return &cache, nil
	}
	apikey := w.APIKey
	url := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%s&q=%f,%f&days=3&aqi=yes&alerts=yes", apikey, location.Latitude, location.Longitude)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := w.HttpClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	w.RedisCache.SetCache(key, sb)
	// log.Printf(sb)
	return &sb, nil
}
