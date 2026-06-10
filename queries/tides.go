package queries

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/storage"
	"github.com/zmaillard/whereami/templates"
)

type tide struct {
	HasTide          bool
	TideStationName  string
	TideStationState string
	StationId        string
	Predictions      []struct {
		T    string `json:"t"`
		V    string `json:"v"`
		Type string `json:"type"`
	} `json:"predictions"`
	StationTimeZoneOffset int
	TimeZone              string
}

func (t tide) SetResults(dto *templates.ResultDto) {
	dto.HasTideStation = t.HasTide
	dto.TideStationName = fmt.Sprintf("%s, %s", t.TideStationName, t.TideStationState)

	var tides []templates.TidePredictions

	now := time.Now()

	for _, prediction := range t.Predictions {
		localTime := time.FixedZone(t.TimeZone, t.StationTimeZoneOffset*60*60)

		v, _ := strconv.ParseFloat(prediction.V, 64)
		timeUtc, _ := time.Parse("2006-01-02 15:04", prediction.T)

		// Get next 4 high/low tides after now
		if timeUtc.After(now) && len(tides) < 4 {
			timeLocal := timeUtc.In(localTime)
			p := templates.TidePredictions{
				Stage:  prediction.Type,
				Height: v,
				Time:   timeLocal,
			}
			tides = append(tides, p)

		}
	}
	dto.TidePredictions = tides
}

func (d *Database) GetTides(client *http.Client, kv storage.Storage) models.Querier {
	return func(coordinates models.Coordinates) (models.Result, error) {
		queryRaw := `select b.id,b.name,b.state, a.distance_m / 1000.0 as dist_km
		from knn2 a JOIN tidestations AS b ON (b.ogc_fid = a.fid)
		where f_table_name = 'tidestations' and ref_geometry = MakePoint(%v,%v) and radius = 1.0 and max_items = 1`

		query := fmt.Sprintf(queryRaw, coordinates.Longitude(), coordinates.Latitude())

		row := d.db.QueryRow(query)
		var station, stationId, state string
		var distance float64
		if err := row.Scan(&stationId, &station, &state, &distance); err != nil {
			return &tide{HasTide: false}, nil
		}

		timezoneOffset, timeZone, err := getTimeZoneOffsetOfStation(client, stationId)
		if err != nil {
			return &tide{HasTide: false}, nil
		}

		today := time.Now()
		startDate := today.Format("20060102")

		url := fmt.Sprintf("https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?begin_date=%v&range=48&station=%s&product=predictions&datum=MLLW&time_zone=lst_ldt&interval=hilo&units=english&application=DataAPI_Sample&format=json", startDate, stationId)
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

		var value tide
		err = json.Unmarshal(body, &value)
		if err != nil {
			fmt.Printf("Error parsing JSON %s", string(body))
			return nil, err
		}
		value.StationId = stationId
		value.TideStationName = station
		value.TideStationState = state
		value.HasTide = true
		value.StationTimeZoneOffset = timezoneOffset
		value.TimeZone = timeZone

		return &value, nil
	}

}

func getTimeZoneOffsetOfStation(client *http.Client, station string) (int, string, error) {
	url := fmt.Sprintf("https://api.tidesandcurrents.noaa.gov/mdapi/prod/webapi/stations/%s.json", station)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return -1, "", err
	}
	req.Header.Add("accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return -1, "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, "", err
	}

	var value stationInfo
	err = json.Unmarshal(body, &value)
	if err != nil {
		fmt.Printf("Error parsing JSON %s", string(body))
		return -1, "", err
	}

	if value.Count == 0 {
		return -1, "", nil
	}

	offset := value.Stations[0].Timezonecorr
	timeZone := value.Stations[0].Timezone
	if value.Stations[0].Observedst {
		offset += 1
		timeZone = strings.Replace(timeZone, "ST", "DT", 1)

	}

	return offset, timeZone, nil

}

type stationInfo struct {
	Count    int         `json:"count"`
	Units    interface{} `json:"units"`
	Stations []struct {
		Tidal      bool   `json:"tidal"`
		Greatlakes bool   `json:"greatlakes"`
		Shefcode   string `json:"shefcode"`
		Details    struct {
			Self string `json:"self"`
		} `json:"details"`
		Sensors struct {
			Self string `json:"self"`
		} `json:"sensors"`
		Floodlevels struct {
			Self string `json:"self"`
		} `json:"floodlevels"`
		Datums struct {
			Self string `json:"self"`
		} `json:"datums"`
		Supersededdatums struct {
			Self string `json:"self"`
		} `json:"supersededdatums"`
		HarmonicConstituents struct {
			Self string `json:"self"`
		} `json:"harmonicConstituents"`
		Benchmarks struct {
			Self string `json:"self"`
		} `json:"benchmarks"`
		TidePredOffsets struct {
			Self string `json:"self"`
		} `json:"tidePredOffsets"`
		OfsMapOffsets struct {
			Self string `json:"self"`
		} `json:"ofsMapOffsets"`
		State        string `json:"state"`
		Timezone     string `json:"timezone"`
		Timezonecorr int    `json:"timezonecorr"`
		Observedst   bool   `json:"observedst"`
		Stormsurge   bool   `json:"stormsurge"`
		Nearby       struct {
			Self string `json:"self"`
		} `json:"nearby"`
		Forecast        bool    `json:"forecast"`
		Outlook         bool    `json:"outlook"`
		HTFhistorical   bool    `json:"HTFhistorical"`
		HTFmonthly      bool    `json:"HTFmonthly"`
		NonNavigational bool    `json:"nonNavigational"`
		Inundationdb    bool    `json:"inundationdb"`
		Id              string  `json:"id"`
		Name            string  `json:"name"`
		Lat             float64 `json:"lat"`
		Lng             float64 `json:"lng"`
		Affiliations    string  `json:"affiliations"`
		Portscode       string  `json:"portscode"`
		Products        struct {
			Self string `json:"self"`
		} `json:"products"`
		Disclaimers struct {
			Self string `json:"self"`
		} `json:"disclaimers"`
		Notices struct {
			Self string `json:"self"`
		} `json:"notices"`
		Self     string `json:"self"`
		Expand   string `json:"expand"`
		TideType string `json:"tideType"`
	} `json:"stations"`
	Self interface{} `json:"self"`
}
