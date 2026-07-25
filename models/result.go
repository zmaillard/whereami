package models

import (
	"github.com/zmaillard/whereami/templates"
)

type Result interface {
	SetResults(dto *templates.ResultDto)
}

type Coordinates interface {
	Latitude() float64
	Longitude() float64
	AsSpatialIndexQueryParameter() (float64, float64, float64, float64)
}
