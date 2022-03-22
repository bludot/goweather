package main

import (
	"context"
	"github.com/bludot/goweather/config"
	"github.com/bludot/goweather/controllers"
	"github.com/bludot/goweather/rediscache"
	"github.com/bludot/goweather/tracing"
	"github.com/bludot/goweather/weatherapi"
	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	healthcheck "github.com/heptiolabs/healthcheck"
)

func setupHealthCheck() http.Handler {
	health := healthcheck.NewHandler()
	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))

	// Our app is not ready if we can't connect to our database (`var db *sql.DB`) in <1s.
	// health.AddReadinessCheck("database", healthcheck.DatabasePingCheck(db, 1*time.Second))
	return health
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	ctx := context.Background()
	log.Println("Starting Service")
	c, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Config loaded", c.Tracing.URL)

	prv, err := tracing.TracerProvider(c)
	if err != nil {
		log.Fatal(err)
	}
	defer prv.Close(ctx)
	m := ginmetrics.GetMonitor()
	m.SetMetricPath("/metrics")
	// +optional set slow time, default 5s
	m.SetSlowTime(10)
	// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
	// used to p95, p99
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})

	cache := rediscache.NewRedisCache(c.RedisDB)
	weatherApi := weatherapi.NewWeatherAPI(c.WeatherAPIConfig, cache)
	ctr := controllers.NewController(cache, weatherApi)
	r := gin.Default()
	// set middleware for gin
	m.Use(r)
	r.POST("/current", CORSMiddleware(), ctr.GetCurrentWeather)
	r.POST("/forecast", CORSMiddleware(), ctr.GetForecast)
	r.Any("/live", gin.WrapH(setupHealthCheck()))
	r.Any("/ready", gin.WrapH(setupHealthCheck()))
	log.Println(strconv.Itoa(c.AppConfig.Port))
	r.Run(":" + strconv.Itoa(c.AppConfig.Port))
	// s := createServer()

	// metrics.SetupMetrics()

	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	//if err := r.Shutdown(ctx); err != nil {
	//	log.Fatal("Server forced to shutdown")
	//}
	log.Println("Server exiting")
}
