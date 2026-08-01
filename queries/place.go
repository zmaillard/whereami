package queries

import (
	"context"
	"log/slog"

	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/util"
)

type place struct {
	Name                    string
	Type                    string
	NextLargestCity         string
	NextLargestCityState    string
	NextLargestCityDistance float64
	NextLargestCityBearing  string
}

func (c place) SetResults(dto *templates.ResultDto) {
	dto.PlaceName = c.Name
	dto.PlaceType = c.Type
	dto.NextLargestCity = c.NextLargestCity
	dto.NextLargestCityState = c.NextLargestCityState
	dto.NextLargestCityDistance = c.NextLargestCityDistance
	dto.NextLargestCityBearing = c.NextLargestCityBearing
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
	var distance, longitude, latitude float64
	var p place
	err = incorpStmt.QueryRow(util.AsSpatialIndexQueryParameter(coords)).Scan(&name, &nextLargestPlaceId)
	if err == nil {
		slog.Info("Found incorporated place", "name", name)

		if nextLargestPlaceId != nil {
			nextIncorpStmt, err := d.db.PrepareContext(ctx, `SELECT  nlip.PLACE_NAME, nlip.STATE_NAME, ST_X(ST_CENTROID(nlip.shape)),ST_Y(ST_CENTROID(nlip.shape)), CvtToUsMi(Distance(ST_POINT(?,?),nlip.shape, 1 )) AS dist_m
			FROM incorporated_place nlip
			where nlip.fid = ?`)
			if err != nil {
				return nil, err
			}
			defer nextIncorpStmt.Close()
			nextIncorpStmt.QueryRow(coords.Longitude(), coords.Latitude(), *nextLargestPlaceId).Scan(&nextLargestPlaceName, &nextLargestPlaceState, &longitude, &latitude, &distance)

			sc := &summitCoords{lng: longitude, lat: latitude}
			bearing := util.GetBearing(coords, sc)
			bearingDir := util.BearingToDirection(bearing)

			return &place{Name: name, Type: "Incorporated", NextLargestCity: nextLargestPlaceName, NextLargestCityState: nextLargestPlaceState, NextLargestCityDistance: distance, NextLargestCityBearing: bearingDir}, nil

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

	err = unincorpStmt.QueryRow(util.AsSpatialIndexQueryParameter(coords)).Scan(&name)
	if err != nil {
		slog.Info("No unincorporated place found", "name", name)
		slog.Error(err.Error())
	}
	slog.Info("Found unincorporated place", "name", name)
	p = place{Name: name, Type: "Unincorporated"}

	findNextPlaceStmt, err := d.db.PrepareContext(ctx, `select b.PLACE_NAME, b.STATE_NAME,ST_X(ST_CENTROID(b.shape)),ST_Y(ST_CENTROID(b.shape)), CvtToUsMi(a.distance_m) as dist_miles
		from knn2 a JOIN incorporated_place AS b ON (b.fid = a.fid)
		where f_table_name = 'incorporated_place' and ref_geometry = MakePoint(?,?) and radius = 1.0 and max_items = 1 AND expand = 1;`)
	if err != nil {
		return nil, err
	}
	defer findNextPlaceStmt.Close()

	err = findNextPlaceStmt.QueryRow(coords.Longitude(), coords.Latitude()).Scan(&p.NextLargestCity, &p.NextLargestCityState, &longitude, &latitude, &p.NextLargestCityDistance)
	if err != nil {
		slog.Info("No Nearest Place Found")
	}
	sc := &summitCoords{lng: longitude, lat: latitude}
	bearing := util.GetBearing(coords, sc)
	bearingDir := util.BearingToDirection(bearing)

	p.NextLargestCityBearing = bearingDir

	return &p, nil
}
