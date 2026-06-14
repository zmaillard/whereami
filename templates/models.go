package templates

import (
	"fmt"
	"strings"
	"time"

	_ "time/tzdata"
)

type ResultDto struct {
	County                   string            `json:"county"`
	State                    string            `json:"state"`
	Latitude                 float64           `json:"latitude"`
	Longitude                float64           `json:"longitude"`
	Place                    string            `json:"place"`
	WeatherServiceOfficeCode string            `json:"weather_service_office_code"`
	WeatherServiceOfficeName string            `json:"weather_service_office_name"`
	Sunrise                  time.Time         `json:"sunrise"`
	Sunset                   time.Time         `json:"sunset"`
	Elevation                float64           `json:"elevation"`
	NearestSummit            string            `json:"nearest_summit"`
	NearestSummitElevation   float64           `json:"nearest_summit_elevation"`
	NearestSummitDistance    float64           `json:"nearest_summit_distance"`
	Tributaries              []string          `json:"tributaries"`
	CurrentHuc               string            `json:"current_huc"`
	TimeZone                 string            `json:"time_zone"`
	HasTideStation           bool              `json:"has_tide_station"`
	TideStationName          string            `json:"tide_station_name"`
	TidePredictions          []TidePredictions `json:"tide_predictions"`
	EcoRegionLevel1          string            `json:"eco_region_level_1"`
	EcoRegionLevel2          string            `json:"eco_region_level_2"`
	EcoRegionLevel3          string            `json:"eco_region_level_3"`
	EcoRegionLevel4          string            `json:"eco_region_level_4"`
	ForecastPeriod1Name      string            `json:"forecast_period_1_name"`
	ForecastPeriod2Name      string            `json:"forecast_period_2_name"`
	ForecastPeriod1Details   string            `json:"forecast_period_1_details"`
	ForecastPeriod2Details   string            `json:"forecast_period_2_details"`
	CongressionalDistrict    string            `json:"congressional_district"`
	PlaceName                string            `json:"place_name"`
	PlaceType                string            `json:"place_type"`
	NextLargestCity          string            `json:"next_largest_city"`
	NextLargestCityState     string            `json:"next_largest_city_state"`
	NationalPark             string            `json:"national_park"`
	NationalParkDistance     float64           `json:"national_park_distance"`
	NextLargestCityDistance  float64           `json:"next_largest_city_distance"`
}

type TidePredictions struct {
	Stage  string
	Height float64
	Time   time.Time
}

func (tp TidePredictions) GetStage() string {
	if tp.Stage == "H" {
		return "High"
	} else if tp.Stage == "L" {
		return "Low"
	}
	return tp.Stage
}

func (d ResultDto) HasEcoRegion() bool {
	if d.EcoRegionLevel1 == "" {
		return false
	}
	return true
}

func (d ResultDto) GetFormattedElevation() string {
	return fmt.Sprintf("%.2f feet", d.Elevation)
}

func (d ResultDto) GetFormattedSummit() string {
	return fmt.Sprintf("%s (%.2f feet - %.2f miles away)", d.NearestSummit, d.NearestSummitElevation, d.NearestSummitDistance)
}

func (d ResultDto) GetFormattedTributaries() string {
	return strings.Join(d.Tributaries, " → ")
}

func (d ResultDto) FormatNearestLargestCity() string {
	return fmt.Sprintf("%s, %s (%.2f miles away)", d.NextLargestCity, d.NextLargestCityState, d.NextLargestCityDistance)
}

func (d ResultDto) GetSink() string {
	if len(d.Tributaries) > 0 {
		return d.Tributaries[len(d.Tributaries)-1]
	}
	return ""
}

func (d ResultDto) GetCoordinates() string {
	return fmt.Sprintf("%.4f, %.4f", d.Latitude, d.Longitude)
}
func (d ResultDto) GetFormattedOfficeLink() string {
	return fmt.Sprintf("https://www.weather.gov/%s", d.WeatherServiceOfficeCode)
}

func (d ResultDto) FormatPlaceName() string {
	if d.PlaceType == "Unincorporated" && d.PlaceName != "" {
		return fmt.Sprintf("%s (Unincorporated)", d.PlaceName)
	}
	return d.PlaceName
}

func (d ResultDto) FormatNationalPark() string {
	return fmt.Sprintf("%s (%.2f miles away)", d.NationalPark, d.NationalParkDistance)
}

func (d ResultDto) FormatTimeZone() string {
	loc, err := time.LoadLocation(d.TimeZone)
	if err != nil {
		return d.TimeZone
	}

	localTime := time.Now().In(loc)

	formattedTime := localTime.Format("2006-01-02 15:04:05")

	return fmt.Sprintf("%s (Local Time: %s)", d.TimeZone, formattedTime)
}

type BaseDto struct {
	VersionNumber string
	MapboxToken   string
}
