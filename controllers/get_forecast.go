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
	ctx, span := tracing.NewSpan(c, method, nil)
	defer span.End()
	var location weatherapi.Location
	err := c.BindJSON(&location)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		span.AddSpanError(err)
		span.FailSpan(err.Error())
		return
	}

	var forecast *string
	if location.Zip != 0 || location.City != "" {
		forecast, err = ctr.WeatherApi.GetForecastByQuery(ctx, &location)
	} else {
		forecast, err = ctr.WeatherApi.GetForecast(ctx, &location)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		span.AddSpanError(err)
		span.FailSpan(err.Error())
		return
	}
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, *forecast)
}
