package queries

import (
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
	stmt, err := d.db.Prepare("SELECT name FROM congressional_district WHERE ST_CONTAINS(geometry, ST_Point(?, ?))")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var name string
	err = stmt.QueryRow(coords.Longitude(), coords.Latitude()).Scan(&name)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("Found congressional district", "name", name)
	return &congressionalDistrict{Name: name}, nil
}
