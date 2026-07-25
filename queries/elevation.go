package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/zmaillard/whereami/metrics"
	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
)

type elevationResponse struct {
	Value float64 `json:"value"`
}

func (er *elevationResponse) SetResults(dto *templates.ResultDto) {
	dto.Elevation = er.Value
}

func (d *Database) GetElevation(client *http.Client) models.Querier {
	return func(ctx context.Context, coordinates models.Coordinates) (models.Result, error) {
		slog.Info("Getting elevation for coordinates", "latitude", coordinates.Latitude(), "longitude", coordinates.Longitude())
		url := fmt.Sprintf("https://epqs.nationalmap.gov/v1/json?x=%v&y=%v&wkid=4326&units=Feet&includeDate=false", coordinates.Longitude(), coordinates.Latitude())
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		req.Header.Add("accept", "application/json")

		// Record API call metrics
		start := time.Now()
		res, err := client.Do(req)
		duration := time.Since(start).Seconds()

		if err != nil {
			slog.Error(err.Error())
			metrics.RecordAPICallError("usgs_elevation")
			metrics.RecordAPICallDuration("usgs_elevation", duration)
			return nil, err
		}
		metrics.RecordAPICallSuccess("usgs_elevation")
		metrics.RecordAPICallDuration("usgs_elevation", duration)

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}

		var value elevationResponse
		err = json.Unmarshal(body, &value)
		if err != nil {
			slog.Error("Error parsing JSON %s", string(body))
			return nil, err
		}

		slog.Info("Found Elevation")
		return d.GetNearestSummit(ctx, coordinates, value.Value)
	}
}
