package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	healthcheck "github.com/heptiolabs/healthcheck"
	env "github.com/joho/godotenv"
)

const envFile = ".env"

var loadEnv = env.Load

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
		forecast, err := getForecast(&location)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(resp, "%s", *forecast)
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
		weather, err := getCurrentWeather(&location)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(resp, "%s", *weather)
		return
	default:
		log.Println("error no 404")
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprint(resp, "not found")
	}
}

func setupHealthCheck() (healthcheckserver *http.Server) {
	health := healthcheck.NewHandler()
	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))
	go http.ListenAndServe("0.0.0.0:8086", health)
	// Our app is not ready if we can't connect to our database (`var db *sql.DB`) in <1s.
	// health.AddReadinessCheck("database", healthcheck.DatabasePingCheck(db, 1*time.Second))
	return
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

	log.Println("Starting Service")
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return
}

func main() {
	restInit()
	s := createServer()
	setupHealthCheck()
	setupMetrics()

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
