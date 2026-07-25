package queries

import (
	"context"
	"log/slog"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/util"
)

type County struct {
	Name      string
	StateName string
	Latitude  float64
	Longitude float64
}

func (c County) SetResults(dto *templates.ResultDto) {
	dto.County = c.Name
	dto.State = c.StateName
	dto.Latitude = c.Latitude
	dto.Longitude = c.Longitude
}

func (d *Database) GetCounty(ctx context.Context, coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting county for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	stmt, err := d.db.PrepareContext(ctx, `SELECT c.namelsad,c.statefp FROM county c WHERE ST_CONTAINS(c.geom, ST_POINT(?,?))
		AND c.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'county' AND search_frame = ST_Point(?,?))`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var name, stfips string
	err = stmt.QueryRow(coords.Longitude(), coords.Latitude(), coords.Longitude(), coords.Latitude()).Scan(&name, &stfips)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("Found county", "name", name, "state_fips", stfips)
	return &County{Name: name, StateName: util.StateFromStateFips(stfips), Longitude: coords.Longitude(), Latitude: coords.Latitude()}, nil
}
