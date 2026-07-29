package queries

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/util"
)

type ecoregions struct {
	Level4USCode string `db:"us_l4code"`
	Level4USName string `db:"us_l4name"`
	Level3USCode string `db:"us_l3code"`
	Level3USName string `db:"us_l3name"`
	Level3NACode string `db:"na_l3code"`
	Level3NAName string `db:"na_l3name"`
	Level2NACode string `db:"na_l2code"`
	Level2NAName string `db:"na_l2name"`
	Level1NACode string `db:"na_l1code"`
	Level1NAName string `db:"na_l1name"`
	Level1Key    string `db:"l1_key"`
	Level2Key    string `db:"l2_key"`
	Level3Key    string `db:"l3_key"`
	Level4Key    string `db:"l4_key"`
}

func (e ecoregions) String() string {
	return fmt.Sprintf("%s|%s|%s|%s", e.Level1Key, e.Level2Key, e.Level3Key, e.Level4Key)
}

func (e ecoregions) SetResults(dto *templates.ResultDto) {
	dto.EcoRegionLevel1 = e.Level1Key
	dto.EcoRegionLevel2 = e.Level2Key
	dto.EcoRegionLevel3 = e.Level3Key
	dto.EcoRegionLevel4 = e.Level4Key
}

func (d *Database) GetEcoregions(ctx context.Context, coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting ecoregions for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	stmt, err := d.db.PrepareContext(ctx, `SELECT e.us_l4code, e.us_l4name, e.us_l3code, e.us_l3name, e.na_l3code, e.na_l3name, e.na_l2code, e.na_l2name, e.na_l1code, e.na_l1name, e.l4_key, e.l3_key, e.l2_key, e.l1_key
	FROM ecoregions e WHERE ST_CONTAINS(e.geometry,ST_POINT(?,?))
	AND e.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'ecoregions' AND search_frame = ST_Point(?,?))`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var level4USCode, us_l4name, us_l3code, us_l3name, na_l3code, na_l3name, na_l2code, na_l2name, na_l1code, na_l1name, l4_key, l3_key, l2_key, l1_key string
	err = stmt.QueryRow(util.AsSpatialIndexQueryParameter(coords)).Scan(&level4USCode, &us_l4name, &us_l3code, &us_l3name, &na_l3code, &na_l3name, &na_l2code, &na_l2name, &na_l1code, &na_l1name, &l4_key, &l3_key, &l2_key, &l1_key)
	if err != nil {
		slog.Error(err.Error())
		return d.getAlaskaEcoregions(ctx, coords)
	}

	eco := ecoregions{
		Level4USCode: level4USCode,
		Level4USName: us_l4name,
		Level3USCode: us_l3code,
		Level3USName: us_l3name,
		Level3NACode: na_l3code,
		Level3NAName: na_l3name,
		Level2NACode: na_l2code,
		Level2NAName: na_l2name,
		Level1NACode: na_l1code,
		Level1NAName: na_l1name,
		Level1Key:    l1_key,
		Level2Key:    l2_key,
		Level3Key:    l3_key,
		Level4Key:    l4_key,
	}

	slog.Info("Found Ecoregions", eco.String())
	return eco, nil
}

func (d *Database) getAlaskaEcoregions(ctx context.Context, coords models.Coordinates) (models.Result, error) {

	slog.Info("Getting alaska ecoregions for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())

	stmt, err := d.db.PrepareContext(ctx, `SELECT e.NA_L1KEY, e.NA_L2KEY, e.NA_L3KEY
	FROM ecoregions_alaska e
	WHERE ST_CONTAINS(geom,ST_POINT(?,?))
	AND e.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'ecoregions_alaska' AND search_frame = ST_Point(?,?))`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var level1key, level2key, level3key string
	err = stmt.QueryRow(util.AsSpatialIndexQueryParameter(coords)).Scan(&level1key, &level2key, &level3key)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	eco := ecoregions{
		Level1Key: level1key,
		Level2Key: level2key,
		Level3Key: level3key,
	}
	slog.Info("Found Alaska Ecoregions", eco.String())
	return eco, nil
}
