package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/zmaillard/whereami/config"
	"github.com/zmaillard/whereami/metrics"
	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
)

//

func (d *Database) GetAQI(client *http.Client, cfg *config.Config) models.Querier {
	return func(ctx context.Context, coordinates models.Coordinates) (models.Result, error) {
		slog.Info("Getting AQI for coordinates", "latitude", coordinates.Latitude(), "longitude", coordinates.Longitude())

		url := fmt.Sprintf("https://www.airnowapi.org/aq/forecast/current/?format=application/json&latitude=%v&longitude=%v&API_KEY=%s", coordinates.Latitude(), coordinates.Longitude(), cfg.AirNowKey)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			slog.Error("Error creating request", "err", err)
			return nil, err
		}
		req.Header.Add("accept", "application/json")

		// Record API call metrics for tide predictions
		start := time.Now()
		res, err := client.Do(req)
		duration := time.Since(start).Seconds()

		if err != nil {
			slog.Error("Error getting response", "err", err)
			metrics.RecordAPICallError("epa_aqi")
			metrics.RecordAPICallDuration("epa_aqi", duration)
			return nil, err
		}
		metrics.RecordAPICallSuccess("epa_aqi")
		metrics.RecordAPICallDuration("epa_eqi", duration)

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			slog.Error("Error reading response", "err", err)
			return nil, err
		}

		var value []aqi
		err = json.Unmarshal(body, &value)
		if err != nil {
			slog.Error("Error parsing JSON %s", string(body))
			return aqiResults(value), nil //No results have valid json - different structure
		}

		slog.Info("Got AQI")
		return aqiResults(value), nil
	}
}

type AQIDate time.Time

func (a *AQIDate) UnmarshalJSON(bytes []byte) error {
	s := strings.Trim(string(bytes), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	*a = AQIDate(t)
	return nil
}

func (a AQIDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(a))
}

type aqi struct {
	DateIssue         AQIDate `json:"dateIssue"`
	DateValid         AQIDate `json:"dateValid"`
	ReportingArea     string  `json:"reportingArea"`
	ReportingAreaCode string  `json:"reportingAreaCode"`
	StateCode         string  `json:"stateCode"`
	ParameterName     string  `json:"parameterName"`
	Value             int     `json:"aqi"`
	ForecastAgency    string  `json:"forecastAgency"`
	CategoryNumber    int     `json:"categoryNumber"`
	CategoryName      string  `json:"categoryName"`
	ActionDay         bool    `json:"actionDay"`
	Discussion        string  `json:"discussion"`
}

type aqiResults []aqi

func (a aqiResults) SetResults(dto *templates.ResultDto) {
	var aqiList []templates.AQI
	for _, r := range a {
		aqiList = append(aqiList, templates.AQI{
			Type:       r.ParameterName,
			Value:      r.Value,
			Category:   r.CategoryName,
			Agency:     r.ForecastAgency,
			DateIssued: time.Time(r.DateIssue),
			DateValid:  time.Time(r.DateValid),
		})
	}

	dto.AQI = aqiList

}
