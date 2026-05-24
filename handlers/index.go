package handlers

import (
	"github.com/labstack/echo/v5"
	"github.com/zmaillard/whereami/config"
	"github.com/zmaillard/whereami/templates"
	"github.com/zmaillard/whereami/templates/components"
	"github.com/zmaillard/whereami/util"
)

func Index(cfg *config.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		return util.RenderView(c, templates.Dashboard(templates.IndexDto{
			BaseDto: components.BaseDto{VersionNumber: "0.1", MapboxToken: cfg.MapboxToken},
		}))
	}
}

func About(cfg *config.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		return util.RenderView(c, templates.About(templates.AboutDto{
			BaseDto: components.BaseDto{VersionNumber: "0.1", MapboxToken: cfg.MapboxToken},
		}))
	}
}

type Coordinates struct {
	Lat float64 `form:"lat"`
	Lng float64 `form:"lng"`
}

func (c *Coordinates) Latitude() float64 {
	return c.Lat
}
func (c *Coordinates) Longitude() float64 {
	return c.Lng
}
