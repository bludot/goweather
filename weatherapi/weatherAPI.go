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
	"net/url"
)

type Location struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	IP        int     `json:"ip"`
	Zip       int     `json:"zip"`
	City      string  `json:"city"`
	Days      int     `json:"days"`
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
	span.Log(fmt.Sprint("query: ", query))
	// The cache key uses the raw query; only the URL form is escaped.
	escapedQuery := url.QueryEscape(query)
	key := queryCacheKey("current", query)
	cache, err := w.RedisCache.GetCache(spanCtx, key)
	if err != nil {
		log.Println("got here")
		// return ""
	}
	if err == nil {
		return &cache, nil
	}
	apikey := w.APIKey

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apikey, escapedQuery)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := w.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
	key := coordCacheKey("current", location.Latitude, location.Longitude)
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
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
	key := forecastCoordCacheKey(location.Latitude, location.Longitude, location.Days)
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

func (w WeatherAPI) GetForecastByQuery(ctx context.Context, location *Location) (res *string, failed error) {
	method := "GetForecastByQuery"
	spanCtx, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()
	var query string
	if location.Zip != 0 {
		query = fmt.Sprintf("%d", location.Zip)
	} else {
		query = location.City
	}
	span.Log(fmt.Sprint("query: ", query))
	// The cache key uses the raw query; only the URL form is escaped.
	escapedQuery := url.QueryEscape(query)
	key := forecastQueryCacheKey(query, location.Days)
	cache, err := w.RedisCache.GetCache(spanCtx, key)
	if err != nil {
		log.Println("got here")
	}
	if err == nil {
		return &cache, nil
	}

	days := normalizeDays(location.Days)

	apikey := w.APIKey
	url := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%s&q=%s&days=%d&aqi=yes&alerts=yes", apikey, escapedQuery, days)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := w.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sb := string(body)
	w.RedisCache.SetCache(spanCtx, key, sb)
	return &sb, nil
}

func (w WeatherAPI) GetForecastAPICall(ctx context.Context, location *Location) (*string, error) {
	method := "GetForecastAPICall"
	_, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()

	days := normalizeDays(location.Days)

	apikey := w.APIKey
	url := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%s&q=%f,%f&days=%d&aqi=yes&alerts=yes", apikey, location.Latitude, location.Longitude, days)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := w.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//Convert the body to type string
	sb := string(body)

	return &sb, nil
}
