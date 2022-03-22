package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/jinzhu/configor"
)

type Config struct {
	RedisDB          RedisDB
	AppConfig        AppConfig
	WeatherAPIConfig WeatherAPIConfig
	Tracing          TracingConfig
}

type AppConfig struct {
	Name    string `env:"CONFIG__APP_CONFIG__NAME" required:"true" default:"goweather"`
	Version string `env:"APP_VERSION" default:"local"`
	Port    int    `env:"CONFIG__APP_CONFIG__PORT" default:"8080"`
	Mode    string `env:"CONFIG__APP_CONFIG__MODE" default:"debug"`
}

type WeatherAPIConfig struct {
	APIKey string `env:"CONFIG__WEATHER_API_CONFIG__API_KEY" required:"true"`
}

type TracingConfig struct {
	URL string `env:"CONFIG__TRACING__URL" required:"false"`
}

type RedisDB struct {
	Host     string `default:"localhost" env:"CONFIG__REDIS_HOST"`
	Password string `required:"false" env:"CONFIG__REDIS_PASS"`
	Port     int    `default:"6379" env:"CONFIG__REDIS_PORT"`
}

func LoadConfig() (*Config, error) {
	var config Config
	log.Println(getEnv())
	err := configor.
		New(&configor.Config{AutoReload: false}).
		Load(&config, fmt.Sprintf("%s/config.%s.json", getConfigLocation(), getEnv()))

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func getConfigLocation() string {
	_, filename, _, _ := runtime.Caller(0)

	return path.Join(path.Dir(filename), "../config")
}

func getEnv() string {
	val := os.Getenv("APP_ENV")
	// todo: check our stage names and align with them
	switch strings.ToLower(val) {
	case "prod":
		return "prod"
	case "staging":
		return "staging"
	case "test":
		return "test"
	case "qa":
		return "qa"
	default:
		return "dev"
	}
}
