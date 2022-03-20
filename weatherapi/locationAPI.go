package weatherapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bludot/goweather/tracing"

	"io/ioutil"
	"log"
	"net/http"
)

// https://api.bigdatacloud.net/data/reverse-geocode-client?latitude=37.42159&longitude=-122.0837&localityLanguage=en

// we just want the city from this endpoint
type City struct {
	City string `json:"city"`
}

func (w WeatherAPI) GetCity(ctx context.Context, location *Location) *City {
	method := "GetCity"
	_, span := tracing.NewSpan(ctx, method, nil)
	defer span.End()
	url := fmt.Sprintf("https://api.bigdatacloud.net/data/reverse-geocode-client?latitude=%flongitude=%f&localityLanguage=en", location.Latitude, location.Longitude)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		tracing.AddSpanError(span, err)
		return nil
	}
	resp, err := w.HttpClient.Do(request)
	if err != nil {
		tracing.AddSpanError(span, err)
		log.Fatalln(err)
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tracing.AddSpanError(span, err)
		log.Fatalln(err)
		return nil
	}
	var city City
	json.Unmarshal([]byte(body), &city)
	return &city
}
