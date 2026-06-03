package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zmaillard/whereami/config"
	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/queries"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/templates/components"
	"github.com/zmaillard/whereami/util"
)

func InitialQuery(cfg *config.Config, database *queries.Database) echo.HandlerFunc {
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
		resultDto.BaseDto = components.BaseDto{VersionNumber: cfg.VersionNumber}

		return util.RenderView(c, templates.Results(resultDto))
	}
}

func Query(database *queries.Database, httpClient *http.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		coords := new(Coordinates)
		var resultDto components.ResultDto
		if err := c.Bind(coords); err != nil {
			return err
		}

		ops := []models.Querier{database.GetElevation, queries.GetWeather, database.GetStream}
		ch := make(chan models.Result, len(ops))
		for _, query := range ops {
			go func() {
				res, err := query(httpClient, coords)
				if err != nil {
					fmt.Println(err)
				}
				ch <- res
			}()
		}

		for range ops {
			res := <-ch
			res.SetResults(&resultDto)
		}

		return c.JSON(http.StatusOK, resultDto)
	}
}
