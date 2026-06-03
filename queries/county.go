package queries

import (
	"fmt"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/util"
)

type County struct {
	Name      string
	StateName string
}

func (d *Database) GetCounty(coords models.Coordinates) (*County, error) {
	query := fmt.Sprintf("SELECT namelsad,statefp FROM county WHERE ST_CONTAINS(geom, %s)", models.GeomStringFromCoordinate(coords))

	row := d.db.QueryRow(query)
	var name, stfips string
	if err := row.Scan(&name, &stfips); err != nil {
		return nil, err
	}
	return &County{Name: name, StateName: util.StateFromStateFips(stfips)}, nil
}
