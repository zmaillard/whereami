package queries

import (
	"fmt"
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

func (d *Database) GetCounty(coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting county for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	query := fmt.Sprintf("SELECT namelsad,statefp FROM county WHERE ST_CONTAINS(geom, %s)", models.GeomStringFromCoordinate(coords))

	row := d.db.QueryRow(query)
	var name, stfips string
	if err := row.Scan(&name, &stfips); err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("Found county", "name", name, "state_fips", stfips)
	return &County{Name: name, StateName: util.StateFromStateFips(stfips), Longitude: coords.Longitude(), Latitude: coords.Latitude()}, nil
}
