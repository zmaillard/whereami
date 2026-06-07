package queries

import (
	"fmt"
	"strings"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
)

type watersheds struct {
	CurrentHuc  string
	Tributaries []string
}

func (w *watersheds) SetResults(dto *templates.ResultDto) {
	dto.Tributaries = w.Tributaries
	dto.CurrentHuc = w.CurrentHuc
}

func (d *Database) GetStream(coords models.Coordinates) (models.Result, error) {
	query := fmt.Sprintf("SELECT huc12 FROM wbdhu12 WHERE ST_CONTAINS(shape,%s)", models.GeomStringFromCoordinate(coords))

	row := d.db.QueryRow(query)
	var huc12 string
	if err := row.Scan(&huc12); err != nil {
		return nil, err
	}

	baseQuery := `select WBDHU12.name
		from hucclosure
		inner join WBDHU12 on childhuc = WBDHU12.huc12
		where parenthuc = ? order by depth;`

	rows, err := d.db.Query(baseQuery, huc12)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tributaries []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		var cleanName string
		if strings.Contains(name, "-") {
			cleanName = strings.Split(name, "-")[1]
		} else {
			cleanName = name
		}

		if len(tributaries) == 0 || tributaries[len(tributaries)-1] != cleanName {
			tributaries = append(tributaries, cleanName)

		}
	}
	return &watersheds{Tributaries: tributaries, CurrentHuc: huc12}, nil

}
