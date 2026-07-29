package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zmaillard/whereami/config"
	"github.com/zmaillard/whereami/models"
	"github.com/zmaillard/whereami/queries"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/util"
)

func Details(database *queries.Database, httpClient *http.Client, cfg *config.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		coords := new(Coordinates)
		var resultDto templates.ResultDto
		if err := c.Bind(coords); err != nil {
			return err
		}

		ctx := c.Request().Context()
		ops := []models.Querier{
			queries.InstrumentQuery("county", database.GetCounty),
			queries.InstrumentQuery("nationalpark", database.GetNationalPark),
			queries.InstrumentQuery("congress", database.GetCongressionalDistrict),
			queries.InstrumentQuery("place", database.GetPlace),
			queries.InstrumentQuery("ecoregion", database.GetEcoregions),
			queries.InstrumentQuery("elevation", database.GetElevation(httpClient)),
			queries.InstrumentQuery("weather", database.GetWeather(httpClient)),
			queries.InstrumentQuery("tides", database.GetTides(httpClient)),
			queries.InstrumentQuery("stream", database.GetStream),
			queries.InstrumentQuery("aqi", database.GetAQI(httpClient, cfg)),
		}
		ch := make(chan models.Result, 1)
		for _, query := range ops {
			go func() {
				res, err := query(ctx, coords)
				if err != nil {
					fmt.Println(err)
				}
				ch <- res
			}()
		}

		for range ops {
			res := <-ch
			if res != nil {
				res.SetResults(&resultDto)
			}
		}

		return util.RenderView(c, templates.Results(resultDto))
	}
}
