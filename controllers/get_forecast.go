package controllers

import (
	"github.com/bludot/goweather/tracing"
	"github.com/bludot/goweather/weatherapi"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (ctr Controller) GetForecast(c *gin.Context) {
	log.Println("got req")
	method := "GetForecast"
	_, span := tracing.NewSpan(c, method, nil)
	defer span.End()
	var location weatherapi.Location
	err := c.BindJSON(&location)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		span.End()
		return
	}
	forecast, err := ctr.WeatherApi.GetForecast(c, &location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		tracing.AddSpanError(span, err)
		return
	}
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, *forecast)
}
