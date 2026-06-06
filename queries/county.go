package queries

import (
	"fmt"
	"net/http"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates/components"
	"github.com/zmaillard/whereami/util"
)

type County struct {
	Name      string
	StateName string
	Latitude  float64
	Longitude float64
}

func (c County) SetResults(dto *components.ResultDto) {
	dto.County = c.Name
	dto.State = c.StateName
	dto.Latitude = c.Latitude
	dto.Longitude = c.Longitude
}

func (d *Database) GetCounty(_ *http.Client, coords models.Coordinates) (models.Result, error) {
	query := fmt.Sprintf("SELECT namelsad,statefp FROM county WHERE ST_CONTAINS(geom, %s)", models.GeomStringFromCoordinate(coords))

	row := d.db.QueryRow(query)
	var name, stfips string
	if err := row.Scan(&name, &stfips); err != nil {
		return nil, err
	}
	return &County{Name: name, StateName: util.StateFromStateFips(stfips), Longitude: coords.Longitude(), Latitude: coords.Latitude()}, nil
}
