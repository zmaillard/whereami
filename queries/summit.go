package queries

import (
	"fmt"

	"github.com/dhconnelly/rtreego"
	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/templates/components"
)

type Summit struct {
	Name             string
	Elevation        float64
	Distance         float64
	CurrentElevation float64
}

func (s *Summit) SetResults(dto *components.ResultDto) {
	dto.NearestSummit = s.Name
	dto.NearestSummitElevation = s.Elevation
	dto.NearestSummitDistance = s.Distance
	dto.Elevation = s.CurrentElevation
}

func (d *Database) GetNearestSummit(coords models.Coordinates, filterElevation float64) (models.Result, error) {
	results := d.rt.NearestNeighbors(1, rtreego.Point{coords.Longitude(), coords.Latitude()}, elevationFilter(filterElevation))
	if len(results) == 0 {
		return nil, nil
	}

	summitIndex := results[0].(*summitIndex)
	query := fmt.Sprintf(`SELECT feature_name, elevation, Distance(geom, %s, 1 ) / 1000.0 AS dist_km 
		FROM gnis
		WHERE feature_id = %v`, models.GeomStringFromCoordinate(coords), summitIndex.Feature_id)

	row := d.db.QueryRow(query)
	var name string
	var elevation, distance float64
	if err := row.Scan(&name, &elevation, &distance); err != nil {
		return nil, err
	}
	return &Summit{Name: name, Elevation: elevation, Distance: distance, CurrentElevation: filterElevation}, nil
}

func (d *Database) LoadSummitTree() error {
	query := fmt.Sprintf("SELECT ST_X(geom), ST_Y(geom), feature_id, elevation FROM gnis where feature_class='Summit' and elevation is not null")

	d.rt = rtreego.NewTree(2, 25, 50)

	rows, err := d.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var x, y, elevation float64
		var feature_id int
		if err := rows.Scan(&x, &y, &feature_id, &elevation); err != nil {
			return err
		}

		d.rt.Insert(&summitIndex{rtreego.Point{x, y}, elevation, feature_id})
	}

	return nil
}

var tol = 0.01

type summitIndex struct {
	Location   rtreego.Point
	Elevation  float64
	Feature_id int
}

func (s *summitIndex) Bounds() rtreego.Rect {
	// define the bounds of s to be a rectangle centered at s.location
	// with side lengths 2 * tol:
	return s.Location.ToRect(tol)
}

func elevationFilter(elevation float64) rtreego.Filter {
	return func(results []rtreego.Spatial, object rtreego.Spatial) (refuse, abort bool) {
		if obj, ok := object.(*summitIndex); ok {
			return obj.Elevation < elevation, false
		}
		return false, false
	}
}
