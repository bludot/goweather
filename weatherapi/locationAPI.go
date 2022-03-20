package weatherapi

import (
	"encoding/json"
	"fmt"

	"io/ioutil"
	"log"
	"net/http"
)

// https://api.bigdatacloud.net/data/reverse-geocode-client?latitude=37.42159&longitude=-122.0837&localityLanguage=en

// we just want the city from this endpoint
type City struct {
	City string `json:"city"`
}

func (w WeatherAPI) GetCity(location *Location) *City {
	url := fmt.Sprintf("https://api.bigdatacloud.net/data/reverse-geocode-client?latitude=%flongitude=%f&localityLanguage=en", location.Latitude, location.Longitude)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	resp, err := w.HttpClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var city City
	json.Unmarshal([]byte(body), &city)
	return &city
}
