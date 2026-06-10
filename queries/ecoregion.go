package queries

import (
	"fmt"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
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

func (e ecoregions) SetResults(dto *templates.ResultDto) {
	dto.EcoRegionLevel1 = e.Level1Key
	dto.EcoRegionLevel2 = e.Level2Key
	dto.EcoRegionLevel3 = e.Level3Key
	dto.EcoRegionLevel4 = e.Level4Key
}

func (d *Database) GetEcoregions(coords models.Coordinates) (models.Result, error) {
	query := fmt.Sprintf("SELECT us_l4code, us_l4name, us_l3code, us_l3name, na_l3code, na_l3name, na_l2code, na_l2name, na_l1code, na_l1name, l4_key, l3_key, l2_key, l1_key FROM ecoregions WHERE ST_CONTAINS(geometry,%s)", models.GeomStringFromCoordinate(coords))

	row := d.db.QueryRow(query)
	var level4USCode, us_l4name, us_l3code, us_l3name, na_l3code, na_l3name, na_l2code, na_l2name, na_l1code, na_l1name, l4_key, l3_key, l2_key, l1_key string
	if err := row.Scan(&level4USCode, &us_l4name, &us_l3code, &us_l3name, &na_l3code, &na_l3name, &na_l2code, &na_l2name, &na_l1code, &na_l1name, &l4_key, &l3_key, &l2_key, &l1_key); err != nil {
		return nil, err
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

	return eco, nil
}
