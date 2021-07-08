package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	env "github.com/joho/godotenv"
)

const envFile = ".env"

var loadEnv = env.Load

type Location struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	IP        int     `json:"ip"`
}

func getEnv(envstring string) string {
	err := loadEnv(envFile)
	if err != nil {
		log.Fatal(err)
	}
	res, exist := os.LookupEnv(envstring)
	if !exist {
		log.Fatal("no " + envstring + " specified")
	}
	return res
}

func getCurrentWeather(location *Location) string {
	key := fmt.Sprintf("current%f,%f", location.Longitude, location.Latitude)
	cache, err := getCache(key)
	if err != nil {
		log.Println("got here")
		// return ""
	}
	if err == nil {
		log.Printf(cache)
		return cache
	}
	apikey := getEnv("APIKEY")
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%f,%f", apikey, location.Latitude, location.Longitude)
	resp, err := http.Get(url)
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
	return sb
}

func getForecast(location *Location) string {
	key := fmt.Sprintf("forecast%f,%f", location.Longitude, location.Latitude)
	cache, err := getCache(key)
	if err != nil {
		log.Println("got here")
		// return ""
	}
	if err == nil {
		log.Printf(cache)
		return cache
	}
	apikey := getEnv("APIKEY")
	url := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%s&q=%f,%f&days=3&aqi=yes&alerts=yes", apikey, location.Latitude, location.Longitude)
	resp, err := http.Get(url)
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
	return sb
}

func addCorsHeader(res http.ResponseWriter) {
	headers := res.Header()
	headers.Add("Access-Control-Allow-Origin", "*")
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")
	headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
	headers.Add("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
}

func handleForcast(resp http.ResponseWriter, req *http.Request) {
	addCorsHeader(resp)
	switch req.Method {
	case http.MethodOptions:
		resp.WriteHeader(http.StatusOK)
		return
	case http.MethodPost:
		log.Println("got req")
		var location Location
		err := json.NewDecoder(req.Body).Decode(&location)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}
		forecast := getForecast(&location)
		resp.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(resp, "%s", forecast)
		return
	default:
		log.Println("error no 404")
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprint(resp, "not found")
	}
}

func handleCurrentWeather(resp http.ResponseWriter, req *http.Request) {
	addCorsHeader(resp)
	switch req.Method {
	case http.MethodOptions:
		resp.WriteHeader(http.StatusOK)
		return
	case http.MethodPost:
		log.Println("got req")
		var location Location
		err := json.NewDecoder(req.Body).Decode(&location)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}
		weather := getCurrentWeather(&location)
		resp.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(resp, "%s", weather)
		return
	default:
		log.Println("error no 404")
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprint(resp, "not found")
	}
}

func createServer() (s *http.Server) {
	port := getEnv("PORT")
	port = fmt.Sprintf(":%s", port)
	mux := http.NewServeMux()

	mux.HandleFunc("/current", handleCurrentWeather)
	mux.HandleFunc("/forecast", handleForcast)

	s = &http.Server{
		Addr:           port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return
}

func main() {
	s := createServer()

	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown")
	}
	log.Println("Server exiting")
}
