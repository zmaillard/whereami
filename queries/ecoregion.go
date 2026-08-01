package queries

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ecoregions struct {
	Level3NACode string `db:"na_l3code"`
	Level3NAName string `db:"na_l3name"`
	Level2NACode string `db:"na_l2code"`
	Level2NAName string `db:"na_l2name"`
	Level1NACode string `db:"na_l1code"`
	Level1NAName string `db:"na_l1name"`
}

func (e ecoregions) String() string {
	return fmt.Sprintf("%s %s|%s %s|%s %s", e.Level1NACode, e.Level1NAName, e.Level2NACode, e.Level2NAName, e.Level3NACode, e.Level3NAName)
}

func (e ecoregions) SetResults(dto *templates.ResultDto) {
	caser := cases.Title(language.AmericanEnglish)

	dto.EcoRegionLevel1 = fmt.Sprintf("%s %s", e.Level1NACode, caser.String(e.Level1NAName))
	dto.EcoRegionLevel2 = fmt.Sprintf("%s %s", e.Level2NACode, caser.String(e.Level2NAName))
	dto.EcoRegionLevel3 = fmt.Sprintf("%s %s", e.Level3NACode, caser.String(e.Level3NAName))
}

func (d *Database) GetEcoregions(ctx context.Context, coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting ecoregions for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	stmt, err := d.db.PrepareContext(ctx, `SELECT  e.na_l3code, e.na_l3name, e.na_l2code, e.na_l2name, e.na_l1code, e.na_l1name
	FROM ecoregions e WHERE ST_CONTAINS(e.geometry,ST_POINT(?,?))
	AND e.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'ecoregions' AND search_frame = ST_Point(?,?))`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var na_l3code, na_l3name, na_l2code, na_l2name, na_l1code, na_l1name string
	err = stmt.QueryRow(util.AsSpatialIndexQueryParameter(coords)).Scan(&na_l3code, &na_l3name, &na_l2code, &na_l2name, &na_l1code, &na_l1name)
	if err != nil {
		slog.Error(err.Error())
		return d.getAlaskaEcoregions(ctx, coords)
	}

	eco := ecoregions{
		Level3NACode: na_l3code,
		Level3NAName: na_l3name,
		Level2NACode: na_l2code,
		Level2NAName: na_l2name,
		Level1NACode: na_l1code,
		Level1NAName: na_l1name,
	}

	slog.Info("Found Ecoregions", eco.String())
	return eco, nil
}

func (d *Database) getAlaskaEcoregions(ctx context.Context, coords models.Coordinates) (models.Result, error) {

	slog.Info("Getting alaska ecoregions for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())

	stmt, err := d.db.PrepareContext(ctx, `SELECT e.NA_L3CODE, e.NA_L3NAME, e.NA_L2CODE, e.NA_L2NAME, e.NA_L1CODE, e.NA_L1NAME
	FROM ecoregions_alaska e
	WHERE ST_CONTAINS(geom,ST_POINT(?,?))
	AND e.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'ecoregions_alaska' AND search_frame = ST_Point(?,?))`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var na_l3code, na_l3name, na_l2code, na_l2name, na_l1code, na_l1name string
	err = stmt.QueryRow(util.AsSpatialIndexQueryParameter(coords)).Scan(&na_l3code, &na_l3name, &na_l2code, &na_l2name, &na_l1code, &na_l1name)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	eco := ecoregions{
		Level3NACode: na_l3code,
		Level3NAName: na_l3name,
		Level2NACode: na_l2code,
		Level2NAName: na_l2name,
		Level1NACode: na_l1code,
		Level1NAName: na_l1name,
	}
	slog.Info("Found Alaska Ecoregions", eco.String())
	return eco, nil
}
