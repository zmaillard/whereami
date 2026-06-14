package queries

import (
	"fmt"
	"log/slog"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
)

type place struct {
	Name                    string
	Type                    string
	NextLargestCity         string
	NextLargestCityState    string
	NextLargestCityDistance float64
}

func (c place) SetResults(dto *templates.ResultDto) {
	dto.PlaceName = c.Name
	dto.PlaceType = c.Type
	dto.NextLargestCity = c.NextLargestCity
	dto.NextLargestCityState = c.NextLargestCityState
	dto.NextLargestCityDistance = c.NextLargestCityDistance
}

func (d *Database) GetPlace(coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting place for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())

	query := fmt.Sprintf(`SELECT ip.PLACE_NAME,  nlip.PLACE_NAME, nlip.STATE_NAME, CvtToUsMi(Distance(ip.shape,nlip.shape, 1 )) AS dist_m
		FROM incorporated_place ip LEFT OUTER JOIN incorporated_place nlip
		ON ip.next_largest_incorporated_place = nlip.fid
		WHERE ST_CONTAINS(ip.shape, %s)`, models.GeomStringFromCoordinate(coords))

	row := d.db.QueryRow(query)
	var name, nextLargestPlaceName, nextLargestPlaceState string
	var distance float64
	var p place
	if err := row.Scan(&name, &nextLargestPlaceName, &nextLargestPlaceState, &distance); err == nil {
		slog.Info("Found incorporated place", "name", name)
		return &place{Name: name, Type: "Incorporated", NextLargestCity: nextLargestPlaceName, NextLargestCityState: nextLargestPlaceState, NextLargestCityDistance: distance}, nil
	}
	slog.Warn("Incorporated place not found.")
	slog.Info("Getting unincorporated place for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	queryUnincporated := fmt.Sprintf("SELECT PLACE_NAME FROM unincorporated_place WHERE ST_CONTAINS(shape, %s)", models.GeomStringFromCoordinate(coords))

	row = d.db.QueryRow(queryUnincporated)
	if err := row.Scan(&name); err != nil {
		slog.Info("No unincorporated place found", "name", name)
		slog.Error(err.Error())
	}
	slog.Info("Found unincorporated place", "name", name)
	p = place{Name: name, Type: "Unincorporated"}

	queryRawFindNext := `select b.PLACE_NAME, b.STATE_NAME, CvtToUsMi(a.distance_m) as dist_miles
		from knn2 a JOIN incorporated_place AS b ON (b.fid = a.fid)
		where f_table_name = 'incorporated_place' and ref_geometry = MakePoint(%v,%v) and radius = 1.0 and max_items = 1 AND expand = 1;`

	queryFindNext := fmt.Sprintf(queryRawFindNext, coords.Longitude(), coords.Latitude())
	nextLargest := d.db.QueryRow(queryFindNext)
	var nextLargestCity, nextLargestCityState string
	var distanceNext float64
	err := nextLargest.Scan(&nextLargestCity, &nextLargestCityState, &distanceNext)
	if err == nil {
		p.NextLargestCity = nextLargestCity
		p.NextLargestCityState = nextLargestCityState
		p.NextLargestCityDistance = distanceNext
	}

	return &p, nil
}
