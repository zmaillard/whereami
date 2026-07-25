package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zmaillard/whereami/queries"
)

func Geocode(database *queries.Database) echo.HandlerFunc {
	return func(c *echo.Context) error {
		coords := new(Coordinates)
		if err := c.Bind(coords); err != nil {
			return err
		}

		res, err := database.ReverseGeocode(c.Request().Context(), coords)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
	}
}
