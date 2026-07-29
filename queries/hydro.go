package queries

import (
	"context"
	"log/slog"
	"strings"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/util"
)

type watersheds struct {
	CurrentHuc  string
	Tributaries []string
}

func (w *watersheds) SetResults(dto *templates.ResultDto) {
	dto.Tributaries = w.Tributaries
	dto.CurrentHuc = w.CurrentHuc
}

func (d *Database) GetStream(ctx context.Context, coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting stream for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	stmt, err := d.db.PrepareContext(ctx, `SELECT h.huc12 FROM wbdhu12 h WHERE ST_CONTAINS(h.shape,ST_POINT(?,?))
		AND h.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'wbdhu12' AND search_frame = ST_Point(?,?))`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var huc12 string
	err = stmt.QueryRow(util.AsSpatialIndexQueryParameter(coords)).Scan(&huc12)
	if err != nil {
		return nil, err
	}

	closureStmt, err := d.db.PrepareContext(ctx, `select WBDHU12.name
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
