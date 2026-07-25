package queries

import (
	"context"
	"log/slog"

	"github.com/zmaillard/whereami/models"
)

type ReverseGeocode struct {
	Place    string   `json:"place"`
	State    string   `json:"state"`
	Distance *float64 `json:"distance"`
}

func (d *Database) ReverseGeocode(ctx context.Context, coordinates models.Coordinates) (*ReverseGeocode, error) {
	slog.Info("Reverse geocoding coordinates", "latitude", coordinates.Latitude(), "longitude", coordinates.Longitude())
	stmt, err := d.db.PrepareContext(ctx, `SELECT p.name,p.state_name FROM place p
	WHERE ST_CONTAINS(geometry, ST_POINT(?,?))
	AND p.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'place' AND search_frame = ST_Point(?,?))`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var name, state string
	err = stmt.QueryRow(coordinates.AsSpatialIndexQueryParameter()).Scan(&name, &state)
	if err != nil {
		slog.Warn(err.Error())
		slog.Warn("Looking for closest place")
		return d.findClosestPlace(ctx, coordinates)
	}

	slog.Info("Found place", "name", name, "state", state)
	return &ReverseGeocode{Place: name, State: state}, nil

}

func (d *Database) findClosestPlace(ctx context.Context, coords models.Coordinates) (*ReverseGeocode, error) {
	stmt, err := d.db.PrepareContext(ctx, `select b.name,b.state_name, a.distance_m / 1000.0 as dist_km
		from knn2 a JOIN place AS b ON (b.ogc_fid = a.fid)
		where f_table_name = 'place' and ref_geometry = MakePoint(?,?) and radius = 1.0 and max_items = 1 AND expand = 1;`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var name, state string
	var distance float64
	err = stmt.QueryRow(coords.Longitude(), coords.Latitude()).Scan(&name, &state, &distance)
	if err != nil {
		slog.Warn(err.Error())
		return nil, err
	}

	slog.Info("Found closest place", "name", name, "state", state)
	return &ReverseGeocode{Place: name, State: state, Distance: &distance}, nil
}
