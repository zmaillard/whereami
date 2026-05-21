package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zmaillard/whereami/db"
	"github.com/zmaillard/whereami/queries"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/templates/components"
	"github.com/zmaillard/whereami/util"
)

func InitialQuery(database *db.Database) echo.HandlerFunc {
	return func(c *echo.Context) error {
		coords := new(Coordinates)
		var resultDto components.InitialResultDto
		if err := c.Bind(coords); err != nil {
			return err
		}

		res, err := database.GetCounty(coords)
		if err != nil {
			return err
		}

		resultDto.County = res.Name
		resultDto.State = res.StateName
		resultDto.Latitude = coords.Lat
		resultDto.Longitude = coords.Lng
		resultDto.BaseDto = components.BaseDto{VersionNumber: "0.1"}
		resultDto.ViewUrls = components.ViewUrls{}

		return util.RenderView(c, templates.Results(resultDto))
	}
}

func Query(database *db.Database, httpClient *http.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		coords := new(Coordinates)
		var resultDto components.ResultDto
		if err := c.Bind(coords); err != nil {
			return err
		}

		ops := []queries.Querier{queries.GetElevation, queries.GetWeather}
		ch := make(chan queries.Result, len(ops))
		for _, query := range ops {
			go func() {
				res, err := query(httpClient, coords)
				if err != nil {
					panic(err)
				}
				ch <- res
			}()
		}

		for range ops {
			res := <-ch
			res.SetResults(&resultDto)
		}

		eleRes, err := database.GetNearestSummit(coords)
		if err != nil {
			return err
		}
		eleRes.SetResults(&resultDto)
		streamRes, err := database.GetStream(coords)
		if err != nil {
			return err
		}
		streamRes.SetResults(&resultDto)

		return c.JSON(http.StatusOK, resultDto)
	}
}
