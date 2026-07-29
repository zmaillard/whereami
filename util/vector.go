package util

import (
	"math"

	"github.com/zmaillard/whereami/models"
)

// https://www.movable-type.co.uk/scripts/latlong.html
func GetBearing(c1 models.Coordinates, c2 models.Coordinates) float64 {
	y := math.Sin(c2.Longitude()-c1.Longitude()) * math.Cos(c2.Latitude())
	x := math.Cos(c1.Latitude())*math.Sin(c2.Latitude()) - math.Sin(c1.Latitude())*math.Cos(c2.Latitude())*math.Cos(c2.Longitude()-c1.Longitude())
	theta := math.Atan2(y, x)

	return math.Mod(theta*180/math.Pi+360.0, 360.0)
}

func AsSpatialIndexQueryParameter(c models.Coordinates) (float64, float64, float64, float64) {
	return c.Longitude(), c.Latitude(), c.Longitude(), c.Latitude()
}

func BearingToDirection(bearing float64) string {
	switch {
	case bearing >= 337.5 || bearing < 22.5:
		return "N"
	case bearing >= 22.5 && bearing < 67.5:
		return "NE"
	case bearing >= 67.5 && bearing < 112.5:
		return "E"
	case bearing >= 112.5 && bearing < 157.5:
		return "SE"
	case bearing >= 157.5 && bearing < 202.5:
		return "S"
	case bearing >= 202.5 && bearing < 247.5:
		return "SW"
	case bearing >= 247.5 && bearing < 292.5:
		return "W"
	case bearing >= 292.5 && bearing < 337.5:
		return "NW"
	default:
		return ""
	}
}
