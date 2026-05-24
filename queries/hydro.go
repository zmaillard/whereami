package queries

import (
	"fmt"
	"strings"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates/components"
)

type watersheds struct {
	CurrentHuc  string
	Tributaries []string
}

func (w *watersheds) SetResults(dto *components.ResultDto) {
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

	baseQuery := `WITH RECURSIVE next_huc(h,r) AS (
					SELECT '%s',1
					UNION
					SELECT WBDHU12.tohuc,r+1
					FROM WBDHU12, next_huc
					WHERE next_huc.h = WBDHU12.huc12)
					SELECT b.name FROM next_huc
					inner join WBDHU12 b on h = b.huc12
					order by r`
	hucQuery := fmt.Sprintf(baseQuery, huc12)
	rows, err := d.db.Query(hucQuery)
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
