package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Location struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	IP        int     `json:"ip"`
}

func getCurrentWeather(location *Location) (res *string, failed error) {
	key := getCity(location).City + "_current"
	// key := fmt.Sprintf("current%f,%f", location.Longitude, location.Latitude)
	cache, err := getCache(key)
	if err != nil {
		log.Println("got here")
		// return ""
	}
	if err == nil {
		return &cache, nil
	}
	apikey := getEnv("APIKEY")
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%f,%f", apikey, location.Latitude, location.Longitude)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := Client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	setCache(key, sb)
	log.Printf(sb)
	return &sb, nil
}

func getForecast(location *Location) (res *string, failed error) {
	key := getCity(location).City + "_forecast"
	// key := fmt.Sprintf("forecast%f,%f", location.Longitude, location.Latitude)
	cache, err := getCache(key)
	if err != nil {
		log.Println("got here")
		// return ""
	}
	if err == nil {
		return &cache, nil
	}
	apikey := getEnv("APIKEY")
	url := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%s&q=%f,%f&days=3&aqi=yes&alerts=yes", apikey, location.Latitude, location.Longitude)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := Client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	setCache(key, sb)
	log.Printf(sb)
	return &sb, nil
}
