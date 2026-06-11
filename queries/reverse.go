package queries

import (
	"fmt"
	"log/slog"

	"github.com/zmaillard/whereami/models"
)

type ReverseGeocode struct {
	Place    string   `json:"place"`
	State    string   `json:"state"`
	Distance *float64 `json:"distance"`
}

func (d *Database) ReverseGeocode(coordinates models.Coordinates) (*ReverseGeocode, error) {
	slog.Info("Reverse geocoding coordinates", "latitude", coordinates.Latitude(), "longitude", coordinates.Longitude())
	query := fmt.Sprintf("SELECT name,state_name FROM place WHERE ST_CONTAINS(geometry, %s)", models.GeomStringFromCoordinate(coordinates))

	row := d.db.QueryRow(query)
	var name, state string
	if err := row.Scan(&name, &state); err != nil {
		slog.Warn(err.Error())
		slog.Warn("Looking for closest place")
		return d.findClosestPlace(coordinates)
	}
	slog.Info("Found place", "name", name, "state", state)
	return &ReverseGeocode{Place: name, State: state}, nil

}

func (d *Database) findClosestPlace(coordinates models.Coordinates) (*ReverseGeocode, error) {
	queryRaw := `select b.name,b.state_name, a.distance_m / 1000.0 as dist_km
		from knn2 a JOIN place AS b ON (b.ogc_fid = a.fid)
		where f_table_name = 'place' and ref_geometry = MakePoint(%v,%v) and radius = 1.0 and max_items = 1 AND expand = 1;`

	query := fmt.Sprintf(queryRaw, coordinates.Longitude(), coordinates.Latitude())

	row := d.db.QueryRow(query)
	var name, state string
	var distance float64
	if err := row.Scan(&name, &state, &distance); err != nil {
		slog.Warn(err.Error())
		return nil, err
	}

	slog.Info("Found closest place", "name", name, "state", state)
	return &ReverseGeocode{Place: name, State: state, Distance: &distance}, nil
}
