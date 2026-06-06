package components

import (
	"fmt"
	"strings"
	"time"
)

type ResultDto struct {
	County                   string    `json:"county"`
	State                    string    `json:"state"`
	Latitude                 float64   `json:"latitude"`
	Longitude                float64   `json:"longitude"`
	Place                    string    `json:"place"`
	WeatherServiceOfficeCode string    `json:"weather_service_office_code"`
	WeatherServiceOfficeName string    `json:"weather_service_office_name"`
	Sunrise                  time.Time `json:"sunrise"`
	Sunset                   time.Time `json:"sunset"`
	Elevation                float64   `json:"elevation"`
	NearestSummit            string    `json:"nearest_summit"`
	NearestSummitElevation   float64   `json:"nearest_summit_elevation"`
	NearestSummitDistance    float64   `json:"nearest_summit_distance"`
	Tributaries              []string  `json:"tributaries"`
	CurrentHuc               string    `json:"current_huc"`
}

func (d ResultDto) GetFormattedElevation() string {
	return fmt.Sprintf("%.2f feet", d.Elevation)
}

func (d ResultDto) GetFormattedSummit() string {
	return fmt.Sprintf("%s (%.2f feet - %.2f km away)", d.NearestSummit, d.NearestSummitElevation, d.NearestSummitDistance)
}

func (d ResultDto) GetFormattedTributaries() string {
	return strings.Join(d.Tributaries, " → ")
}

func (d ResultDto) GetSink() string {
	if len(d.Tributaries) > 0 {
		return d.Tributaries[len(d.Tributaries)-1]
	}
	return ""
}

func (d ResultDto) GetLatitude() string {
	return fmt.Sprintf("%.4f", d.Latitude)
}
func (d ResultDto) GetLongitude() string {
	return fmt.Sprintf("%.4f", d.Longitude)
}

func (d ResultDto) GetFormattedOfficeLink() string {
	return fmt.Sprintf("https://www.weather.gov/%s", d.WeatherServiceOfficeCode)
}

type BaseDto struct {
	VersionNumber string
	MapboxToken   string
}
