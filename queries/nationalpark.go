package queries

import (
	"context"
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

func (d *Database) GetNationalPark(ctx context.Context, coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting nearest national park for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())
	stmt, err := d.db.PrepareContext(ctx, `select b.name, CvtToUsMi(a.distance_m) as dist_miles
		from knn2 a JOIN national_park_service AS b ON (b.fid = a.fid)
		where f_table_name = 'national_park_service' and ref_geometry = MakePoint(?,?) and radius = 2.0 and max_items = 1 AND expand = 1;`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var name string
	var distance float64
	err = stmt.QueryRow(coords.Longitude(), coords.Latitude()).Scan(&name, &distance)
	if err != nil {
		return nil, err
	}

	slog.Info("Found nearest national park", "name", name)
	return &nationalPark{Name: name, Distance: distance}, nil
}
