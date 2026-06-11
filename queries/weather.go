package queries

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
)

type metadataResult struct {
	OfficeName string
	Context    []interface{} `json:"@context"`
	Id         string        `json:"id"`
	Type       string        `json:"type"`
	Geometry   struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties struct {
		Id                  string `json:"@id"`
		Type                string `json:"@type"`
		Cwa                 string `json:"cwa"`
		Type1               string `json:"type"`
		ForecastOffice      string `json:"forecastOffice"`
		GridId              string `json:"gridId"`
		GridX               int    `json:"gridX"`
		GridY               int    `json:"gridY"`
		Forecast            string `json:"forecast"`
		ForecastHourly      string `json:"forecastHourly"`
		ForecastGridData    string `json:"forecastGridData"`
		ObservationStations string `json:"observationStations"`
		RelativeLocation    struct {
			Type     string `json:"type"`
			Geometry struct {
				Type        string    `json:"type"`
				Coordinates []float64 `json:"coordinates"`
			} `json:"geometry"`
			Properties struct {
				City     string `json:"city"`
				State    string `json:"state"`
				Distance struct {
					UnitCode string  `json:"unitCode"`
					Value    float64 `json:"value"`
				} `json:"distance"`
				Bearing struct {
					UnitCode string `json:"unitCode"`
					Value    int    `json:"value"`
				} `json:"bearing"`
			} `json:"properties"`
		} `json:"relativeLocation"`
		ForecastZone     string `json:"forecastZone"`
		County           string `json:"county"`
		FireWeatherZone  string `json:"fireWeatherZone"`
		TimeZone         string `json:"timeZone"`
		RadarStation     string `json:"radarStation"`
		AstronomicalData struct {
			Sunrise                   time.Time `json:"sunrise"`
			Sunset                    time.Time `json:"sunset"`
			Transit                   time.Time `json:"transit"`
			CivilTwilightBegin        time.Time `json:"civilTwilightBegin"`
			CivilTwilightEnd          time.Time `json:"civilTwilightEnd"`
			NauticalTwilightBegin     time.Time `json:"nauticalTwilightBegin"`
			NauticalTwilightEnd       time.Time `json:"nauticalTwilightEnd"`
			AstronomicalTwilightBegin time.Time `json:"astronomicalTwilightBegin"`
			AstronomicalTwilightEnd   time.Time `json:"astronomicalTwilightEnd"`
		} `json:"astronomicalData"`
		Nwr struct {
			Transmitter    interface{} `json:"transmitter"`
			SameCode       string      `json:"sameCode"`
			AreaBroadcast  interface{} `json:"areaBroadcast"`
			PointBroadcast string      `json:"pointBroadcast"`
		} `json:"nwr"`
	} `json:"properties"`
}

func (mr *metadataResult) SetResults(dto *templates.ResultDto) {
	dto.WeatherServiceOfficeName = mr.OfficeName
	dto.WeatherServiceOfficeCode = mr.Properties.Cwa
	dto.Sunrise = mr.Properties.AstronomicalData.Sunrise
	dto.Sunset = mr.Properties.AstronomicalData.Sunset
	dto.TimeZone = mr.Properties.TimeZone
}

func (d *Database) GetWeather(client *http.Client) models.Querier {
	return func(coordinates models.Coordinates) (models.Result, error) {
		slog.Info("Getting weather metadata for coordinates", "latitude", coordinates.Latitude(), "longitude", coordinates.Longitude())
		url := fmt.Sprintf("https://api.weather.gov/points/%v,%v", coordinates.Latitude(), coordinates.Longitude())
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			slog.Error("Error creating request", "err", err)
			return nil, err
		}
		req.Header.Add("accept", "application/json")
		res, err := client.Do(req)
		if err != nil {
			slog.Error("Error getting weather metadata", "err", err)
			return nil, err
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			slog.Error("Error getting weather metadata", "err", err)
			return nil, err
		}

		var value metadataResult
		err = json.Unmarshal(body, &value)
		if err != nil {
			slog.Error("Error parsing JSON %s", string(body))
			return nil, err
		}

		query := d.db.QueryRow("SELECT name FROM nwsoffice WHERE code = ?", value.Properties.Cwa)
		var officeName string
		err = query.Scan(&officeName)
		if err == nil {
			value.OfficeName = officeName
		} else {
			slog.Warn("Error getting office name from database, using code instead", "err", err)
		}

		return &value, nil
	}

}
