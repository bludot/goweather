package weatherapi

import (
	"context"
	"fmt"
	"github.com/bludot/goweather/config"
	"github.com/bludot/goweather/http_client"
	"github.com/bludot/goweather/rediscache"
	"github.com/bludot/goweather/tracing"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Location struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	IP        int     `json:"ip"`
	Zip       int     `json:"zip"`
	City      string  `json:"city"`
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

func (w WeatherAPI) GetCurrentWeatherByQuery(ctx context.Context, location *Location) (res *string, failed error) {
	method := "GetCurrentWeatherByZipCode"
	spanCtx, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()
	var query string
	if location.Zip != 0 {
		// int to string
		query = fmt.Sprintf("%d", location.Zip)
	} else {
		query = location.City
	}
	// replace spaces with %20
	query = strings.ReplaceAll(query, " ", "%20")
	span.Log(fmt.Sprint("query: ", query))
	key := fmt.Sprintf("current%d", query)
	cache, err := w.RedisCache.GetCache(spanCtx, key)
	if err != nil {
		log.Println("got here")
		// return ""
	}
	if err == nil {
		return &cache, nil
	}
	apikey := w.APIKey

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apikey, query)
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
	w.RedisCache.SetCache(spanCtx, key, sb)

	// log.Printf(sb)
	return &sb, nil
}

func (w WeatherAPI) GetCurrentWeather(ctx context.Context, location *Location) (res *string, failed error) {
	method := "GetCurrentWeather"
	spanCtx, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()
	key := w.GetCity(ctx, location).City + "_current"
	// key := fmt.Sprintf("current%f,%f", location.Longitude, location.Latitude)
	cache, err := w.RedisCache.GetCache(spanCtx, key)
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
	w.RedisCache.SetCache(spanCtx, key, sb)
	// log.Printf(sb)
	return &sb, nil
}

func (w WeatherAPI) GetForecast(ctx context.Context, location *Location) (res *string, failed error) {
	method := "GetForecast"
	spanCtx, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()
	key := w.GetCity(spanCtx, location).City + "_forecast"
	// key := fmt.Sprintf("forecast%f,%f", location.Longitude, location.Latitude)
	cache, err := w.RedisCache.GetCache(spanCtx, key)
	if err != nil {
		log.Println("got here")
		// return ""
	}
	if err == nil {
		return &cache, nil
	}
	sb, err := w.GetForecastAPICall(spanCtx, location)
	if err != nil {
		return nil, err
	}
	w.RedisCache.SetCache(spanCtx, key, *sb)
	// log.Printf(sb)
	return sb, nil
}

func (w WeatherAPI) GetForecastAPICall(ctx context.Context, location *Location) (*string, error) {
	method := "GetForecastAPICall"
	_, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()

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

	return &sb, nil
}
