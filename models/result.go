package models

import (
	"fmt"

	"github.com/zmaillard/whereami/templates"
)

type Result interface {
	SetResults(dto *templates.ResultDto)
}

type Coordinates interface {
	Latitude() float64
	Longitude() float64
}

func GeomStringFromCoordinate(c Coordinates) string {
	return fmt.Sprintf("ST_PointFromText('POINT(%v %v)')", c.Longitude(), c.Latitude())
}
