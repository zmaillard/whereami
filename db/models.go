package db

import (
	"fmt"

	"github.com/zmaillard/whereami/queries"
)

type County struct {
	Name      string
	StateName string
}

func GeomStringFromCoordinate(c queries.Coordinates) string {
	return fmt.Sprintf("ST_PointFromText('POINT(%v %v)')", c.Longitude(), c.Latitude())
}
