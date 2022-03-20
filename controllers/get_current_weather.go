package controllers

import (
	"github.com/bludot/goweather/weatherapi"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (ctr Controller) GetCurrentWeather(c *gin.Context) {
	var location weatherapi.Location
	err := c.BindJSON(&location)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	forecast, err := ctr.WeatherApi.GetForecast(&location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, *forecast)
}
