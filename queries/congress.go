package queries

import (
	"context"
	"log/slog"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
)

type congressionalDistrict struct {
	Name string
}

func (c congressionalDistrict) SetResults(dto *templates.ResultDto) {
	dto.CongressionalDistrict = c.Name
}

func (d *Database) GetCongressionalDistrict(ctx context.Context, coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting congressional district for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	stmt, err := d.db.PrepareContext(ctx, `SELECT cp.name FROM congressional_district cp
	WHERE ST_CONTAINS(cp.geometry, ST_Point(?,?))
	AND cp.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'congressional_district' AND search_frame = ST_Point(?,?))`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var name string
	err = stmt.QueryRow(coords.Longitude(), coords.Latitude(), coords.Longitude(), coords.Latitude()).Scan(&name)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("Found congressional district", "name", name)
	return &congressionalDistrict{Name: name}, nil
}
