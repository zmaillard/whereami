package queries

import (
	"fmt"
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

func (d *Database) GetCongressionalDistrict(coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting congressional district for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	query := fmt.Sprintf("SELECT name FROM congressional_district WHERE ST_CONTAINS(geometry, %s)", models.GeomStringFromCoordinate(coords))

	row := d.db.QueryRow(query)
	var name string
	if err := row.Scan(&name); err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("Found congressional district", "name", name)
	return &congressionalDistrict{Name: name}, nil
}
