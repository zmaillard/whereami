package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/queries"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/util"
)

func Details(database *queries.Database, httpClient *http.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		coords := new(Coordinates)
		var resultDto templates.ResultDto
		if err := c.Bind(coords); err != nil {
			return err
		}

		ops := []models.Querier{database.GetCounty, database.GetElevation(httpClient), database.GetWeather(httpClient), database.GetStream}
		ch := make(chan models.Result, len(ops))
		for _, query := range ops {
			go func() {
				res, err := query(coords)
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

		return util.RenderView(c, templates.Results(resultDto))
	}
}
