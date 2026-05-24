package queries

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates/components"
)

type elevationResponse struct {
	Value float64 `json:"value"`
}

func (er *elevationResponse) SetResults(dto *components.ResultDto) {
	dto.Elevation = er.Value
}

func (d *Database) GetElevation(client *http.Client, coordinates models.Coordinates) (models.Result, error) {
	url := fmt.Sprintf("https://epqs.nationalmap.gov/v1/json?x=%v&y=%v&wkid=4326&units=Feet&includeDate=false", coordinates.Longitude(), coordinates.Latitude())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var value elevationResponse
	err = json.Unmarshal(body, &value)
	if err != nil {
		fmt.Printf("Error parsing JSON %s", string(body))
		return nil, err
	}

	return d.GetNearestSummit(coordinates, value.Value)
}
