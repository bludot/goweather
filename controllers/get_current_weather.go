package controllers

import (
	"github.com/bludot/goweather/tracing"
	"github.com/bludot/goweather/weatherapi"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (ctr Controller) GetCurrentWeather(c *gin.Context) {
	method := "GetCurrentWeather"
	_, span := tracing.NewSpan(c, method, nil)
	defer span.End()
	var location weatherapi.Location
	err := c.BindJSON(&location)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		span.AddSpanError(err)
		span.FailSpan(err.Error())
		return
	}

	span.AddSpanEvents("location", map[string]string{
		"event": "location",
	})
	forecast, err := ctr.WeatherApi.GetForecast(c, &location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		span.AddSpanError(err)
		span.FailSpan(err.Error())
		return
	}
	span.Log(*forecast)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, *forecast)
}
