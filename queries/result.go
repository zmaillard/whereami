package queries

import (
	"net/http"

	"github.com/zmaillard/whereami/templates/components"
)

type Querier func(client *http.Client, coordinates Coordinates) (Result, error)

type Result interface {
	SetResults(dto *components.ResultDto)
}
type Coordinates interface {
	Latitude() float64
	Longitude() float64
}
