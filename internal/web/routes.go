package web

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (server *GoRunnerWebServer) setRoutes() {
	// general api
	server.echo.POST("/api/apps", server.registerApp)
	server.echo.GET("/api/health", server.health)

	// per app api
	server.echo.GET("/api/:app", server.appStatus)
	server.echo.PUT("/api/:app", server.updateApp)
	server.echo.DELETE("/api/:app", server.deleteApp)

	// access app
	server.echo.Any("/:app/*", server.proxyRequest)
}

func (server *GoRunnerWebServer) deleteApp(c echo.Context) error {
	appName := c.Param("app")
	app, err := server.runner.GetApp(appName)
	if err != nil {
		return c.JSON(http.StatusNotFound, errStatus{
			app, err,
		})
	}

	err = app.Stop()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			app, err,
		})
	}

	err = app.Purge()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			app, err,
		})
	}

	server.runner.DeleteApp(appName)

	return c.JSON(http.StatusOK, app)
}

func (server *GoRunnerWebServer) appStatus(c echo.Context) error {
	appName := c.Param("app")
	server.logger.Debug().Msgf("get app status - appName=%s", appName)

	goapp, err := server.runner.GetApp(appName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	return c.JSON(http.StatusOK, goapp)
}

func (server *GoRunnerWebServer) registerApp(c echo.Context) error {
	server.logger.Info().Msg("new app...")
	params := new(DeployAppParams)
	err := c.Bind(params)
	if err != nil {
		server.logger.Err(err).Msg("malformed request")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		server.logger.Err(err).Msg("invalid request params")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	server.logger.Info().Msgf("registering app... - app=%s, gitUrl=%s", params.App, params.GitUrl)
	goapp, _ := server.runner.GetApp(params.App)
	if goapp != nil {
		return c.JSON(http.StatusBadRequest, errStatus{
			goapp, errors.New("app already exists"),
		})
	}

	goapp, err = server.runner.RegisterApp(params.App, params.GitUrl)
	if err != nil {
		server.logger.Err(err).Msgf("error registering app. - app=%s, gitUrl=%s", params.App, params.GitUrl)
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	return server.deployApp(c, goapp)
}

func (server *GoRunnerWebServer) updateApp(c echo.Context) error {
	server.logger.Info().Msg("updating app...")
	params := new(UpdateAppParams)
	err := c.Bind(params)
	if err != nil {
		server.logger.Err(err).Msg("malformed request")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		server.logger.Err(err).Msg("invalid request params")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	app, err := server.runner.GetApp(params.App)
	if err != nil {
		server.logger.Err(err).Msgf("app not found. app=%s", params.App)
		return c.JSON(http.StatusNotFound, errStatus{
			nil, err,
		})
	}

	switch params.Action {
	case "deploy":
		return server.deployApp(c, app)
	case "restart":
		return server.restartApp(c, app)
	}

	err = errors.New("unknown command")
	server.logger.Err(err).Msgf("expected: deploy|restart. action=%s", params.Action)
	return c.JSON(http.StatusInternalServerError, errStatus{
		nil, err,
	})
}
