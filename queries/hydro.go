package queries

import (
	"log/slog"
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
	slog.Info("Getting stream for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	stmt, err := d.db.Prepare("SELECT huc12 FROM wbdhu12 WHERE ST_CONTAINS(shape,ST_POINT(?,?))")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var huc12 string
	err = stmt.QueryRow(coords.Longitude(), coords.Latitude()).Scan(&huc12)
	if err != nil {
		return nil, err
	}

	closureStmt, err := d.db.Prepare(`select WBDHU12.name
		from hucclosure
		inner join WBDHU12 on childhuc = WBDHU12.huc12
		where parenthuc = ? order by depth;`)
	if err != nil {
		return nil, err
	}
	defer closureStmt.Close()

	rows, err := closureStmt.Query(huc12)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tributaries []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			slog.Error(err.Error())
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
	slog.Info("Found stream", "huc12", huc12)
	return &watersheds{Tributaries: tributaries, CurrentHuc: huc12}, nil

}
