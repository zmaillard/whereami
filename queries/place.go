package queries

import (
	"context"
	"log/slog"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
)

type place struct {
	Name                    string
	Type                    string
	NextLargestCity         string
	NextLargestCityState    string
	NextLargestCityDistance float64
}

func (c place) SetResults(dto *templates.ResultDto) {
	dto.PlaceName = c.Name
	dto.PlaceType = c.Type
	dto.NextLargestCity = c.NextLargestCity
	dto.NextLargestCityState = c.NextLargestCityState
	dto.NextLargestCityDistance = c.NextLargestCityDistance
}

func (d *Database) GetPlace(ctx context.Context, coords models.Coordinates) (models.Result, error) {
	slog.Info("Getting place for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())

	incorpStmt, err := d.db.PrepareContext(ctx, `SELECT ip.PLACE_NAME,ip.next_largest_incorporated_place
		FROM incorporated_place ip 
		WHERE ST_CONTAINS(ip.shape, ST_POINT(?,?))
		AND ip.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'incorporated_place' AND search_frame = ST_Point(?,?))`)
	if err != nil {
		return nil, err
	}
	defer incorpStmt.Close()

	var name, nextLargestPlaceName, nextLargestPlaceState string
	var nextLargestPlaceId *int
	var distance float64
	var p place
	err = incorpStmt.QueryRow(coords.AsSpatialIndexQueryParameter()).Scan(&name, &nextLargestPlaceId) //, &nextLargestPlaceName, &nextLargestPlaceState, &distance)
	if err == nil {
		slog.Info("Found incorporated place", "name", name)

		if nextLargestPlaceId != nil {
			nextIncorpStmt, err := d.db.PrepareContext(ctx, `SELECT  nlip.PLACE_NAME, nlip.STATE_NAME, CvtToUsMi(Distance(ST_POINT(?,?),nlip.shape, 1 )) AS dist_m
			FROM incorporated_place nlip
			where nlip.fid = ?`)
			if err != nil {
				return nil, err
			}
			defer nextIncorpStmt.Close()
			nextIncorpStmt.QueryRow(coords.Longitude(), coords.Latitude(), *nextLargestPlaceId).Scan(&nextLargestPlaceName, &nextLargestPlaceState, &distance)
			return &place{Name: name, Type: "Incorporated", NextLargestCity: nextLargestPlaceName, NextLargestCityState: nextLargestPlaceState, NextLargestCityDistance: distance}, nil

		}

		return &place{Name: name, Type: "Incorporated"}, nil
	}
	slog.Warn("Incorporated place not found.")
	slog.Info("Getting unincorporated place for coordinates", "latitude", coords.Latitude(), "longitude", coords.Longitude())

	unincorpStmt, err := d.db.PrepareContext(ctx, `SELECT PLACE_NAME FROM unincorporated_place up
	WHERE ST_CONTAINS(up.shape, ST_POINT(?,?))
	AND up.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'unincorporated_place' AND search_frame = ST_Point(?,?))`)
	if err != nil {
		return nil, err
	}
	defer unincorpStmt.Close()

	err = unincorpStmt.QueryRow(coords.AsSpatialIndexQueryParameter()).Scan(&name)
	if err != nil {
		slog.Info("No unincorporated place found", "name", name)
		slog.Error(err.Error())
	}
	slog.Info("Found unincorporated place", "name", name)
	p = place{Name: name, Type: "Unincorporated"}

	findNextPlaceStmt, err := d.db.PrepareContext(ctx, `select b.PLACE_NAME, b.STATE_NAME, CvtToUsMi(a.distance_m) as dist_miles
		from knn2 a JOIN incorporated_place AS b ON (b.fid = a.fid)
		where f_table_name = 'incorporated_place' and ref_geometry = MakePoint(?,?) and radius = 1.0 and max_items = 1 AND expand = 1;`)
	if err != nil {
		return nil, err
	}
	defer findNextPlaceStmt.Close()

	err = findNextPlaceStmt.QueryRow(coords.Longitude(), coords.Latitude()).Scan(&p.NextLargestCity, &p.NextLargestCityState, &p.NextLargestCityDistance)
	if err != nil {
		slog.Info("No Nearest Place Found")
	}

	return &p, nil
}
