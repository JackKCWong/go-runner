package app

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (r *GoRunner) Start() error {
	r.server.GET("/api/:app", r.appStatus)
	r.server.POST("/api/:app", r.deployApp)
	r.server.DELETE("/api/:app", r.deleteApp)

	r.server.Any("/:app", r.proxyRequest)

	return r.server.Start(r.address)
}

func (r *GoRunner) deleteApp(c echo.Context) error {
	appName := c.Param("app")
	app, err := r.getApp(appName)
	if err != nil {
		return err
	}

	app.Stop()

	return c.JSON(http.StatusOK, app)
}

func (r *GoRunner) appStatus(c echo.Context) error {
	appName := c.Param("app")
	app, err := r.getApp(appName)
	if err != nil {
		return c.JSON(http.StatusNotFound, nil)
	}

	return c.JSON(http.StatusOK, app)
}

func (r *GoRunner) deployApp(c echo.Context) error {
	appName := c.Param("app")
	gitUrl := c.Param("gitUrl")
	app, err := r.registerApp(appName, gitUrl)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("%q", err))
	}

	return c.JSON(http.StatusOK, app)
}

func (r *GoRunner) proxyRequest(c echo.Context) error {
	appName := c.Param("app")
	app, err := r.getApp(appName)
	if err != nil {
		return err
	}

	resp, err := app.Handle(c.Request())
	if err != nil {
		return err
	}

	err = resp.Write(c.Response().Writer)
	if err != nil {
		return err
	}

	return nil
}
