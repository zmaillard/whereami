package components

import (
	"time"

	"github.com/labstack/echo/v5"
)

type InitialResultDto struct {
	BaseDto
	ViewUrls
	County    string  `json:"county"`
	State     string  `json:"state"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ResultDto struct {
	WeatherServiceOffice   string    `json:"weather_service_office"`
	Sunrise                time.Time `json:"sunrise"`
	Sunset                 time.Time `json:"sunset"`
	Elevation              float64   `json:"elevation"`
	NearestSummit          string    `json:"nearest_summit"`
	NearestSummitElevation float64   `json:"nearest_summit_elevation"`
	NearestSummitDistance  float64   `json:"nearest_summit_distance"`
	Tributaries            []string  `json:"tributaries"`
	CurrentHuc             string    `json:"current_huc"`
}

type BaseDto struct {
	VersionNumber string
	MapboxToken   string
}

func NewViewUrls(baseUrl string, e *echo.Echo) ViewUrls {
	return ViewUrls{
		BaseUrl: baseUrl,
		echo:    e,
	}
}

type ViewUrls struct {
	BaseUrl string
	echo    *echo.Echo
}
