package queries

import (
	"fmt"
	"log/slog"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
)

type nationalPark struct {
	Name     string
	Distance float64
}

func (c nationalPark) SetResults(dto *templates.ResultDto) {
	dto.NationalPark = c.Name
	dto.NationalParkDistance = c.Distance
}

func (d *Database) GetNationalPark(coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting nearest national park for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	queryRaw := `select b.name, CvtToUsMi(a.distance_m) as dist_miles
		from knn2 a JOIN national_park_service AS b ON (b.fid = a.fid)
		where f_table_name = 'national_park_service' and ref_geometry = MakePoint(%v,%v) and radius = 2.0 and max_items = 1 AND expand = 1;`

	query := fmt.Sprintf(queryRaw, coords.Longitude(), coords.Latitude())

	row := d.db.QueryRow(query)
	var name string
	var distance float64
	if err := row.Scan(&name, &distance); err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("Found nearest national park", "name", name)
	return &nationalPark{Name: name, Distance: distance}, nil
}
