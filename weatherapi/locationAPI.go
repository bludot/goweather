package weatherapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bludot/goweather/tracing"

	"io/ioutil"
	"net/http"
)

// https://api.bigdatacloud.net/data/reverse-geocode-client?latitude=37.42159&longitude=-122.0837&localityLanguage=en

// we just want the city from this endpoint
type City struct {
	City string `json:"city"`
}

// GetCity reverse-geocodes a coordinate pair to a city name.
//
// It returns nil when the lookup cannot be completed, and a City whose City
// field is "" when the coordinates have no associated city (oceans, remote
// areas).  Callers MUST handle both — dereferencing the result directly, or
// using an empty City as a cache key, serves one location's data for another.
// See coordCacheKey in cachekey.go.
func (w WeatherAPI) GetCity(ctx context.Context, location *Location) *City {
	method := "GetCity"
	_, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()
	span.Log(fmt.Sprint("longitude: ", location.Longitude, " latitude: ", location.Latitude))
	url := fmt.Sprintf("https://api.bigdatacloud.net/data/reverse-geocode-client?latitude=%f&longitude=%f&localityLanguage=en", location.Latitude, location.Longitude)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		span.AddSpanError(err)
		span.Log(err.Error())
		return nil
	}
	resp, err := w.HttpClient.Do(request)
	if err != nil {
		span.AddSpanError(err)
		span.Log(err.Error())
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		span.AddSpanError(err)
		span.Log(err.Error())
		return nil
	}
	var city City
	if err := json.Unmarshal(body, &city); err != nil {
		// An unmarshal failure (an error page, a rate-limit body) previously
		// fell through and yielded a City with an empty name.
		span.AddSpanError(err)
		span.Log(err.Error())
		return nil
	}
	span.Log("city: " + city.City)
	return &city
}
