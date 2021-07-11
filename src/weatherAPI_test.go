package main

import (
	"bytes"
	"floretos/weather/src/mocks"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestGetCurrentWeather(t *testing.T) {
	loadEnv = func(filename ...string) (err error) {
		os.Setenv("APIKEY", "notakey")
		os.Setenv("REDIS_HOST", "redis")
		return
	}

	Client = &mocks.MockClient{}
	json := `{"city": "NotACity"}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	// restClient.Clientmocks
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	location := Location{
		Longitude: 100.502762,
		Latitude:  13.756331,
	}
	weather, err := getCurrentWeather(&location)
	if err != nil {
		t.Fail()
	}
	log.Println(weather)
}

func TestGetForecast(t *testing.T) {
	loadEnv = func(filename ...string) (err error) {
		os.Setenv("APIKEY", "notakey")
		os.Setenv("REDIS_HOST", "redis")
		return
	}

	Client = &mocks.MockClient{}
	json := `{"city": "NotACity"}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	// restClient.Clientmocks
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	location := Location{
		Longitude: 100.502762,
		Latitude:  13.756331,
	}
	weather, err := getForecast(&location)
	if err != nil {
		t.Fail()
	}
	log.Println(weather)
}
