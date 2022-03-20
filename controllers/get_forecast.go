package controllers

import (
	"github.com/bludot/goweather/weatherapi"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (ctr Controller) GetForecast(c *gin.Context) {
	log.Println("got req")
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
	c.JSON(http.StatusOK, *forecast)
	return
}
