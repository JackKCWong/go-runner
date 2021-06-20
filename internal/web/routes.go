package web

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (runner *GoRunnerWebServer) setRoutes() {
	runner.echo.GET("/api/:app", runner.appStatus)
	runner.echo.POST("/api/:app", runner.registerApp)
	runner.echo.PUT("/api/:app", runner.updateApp)
	runner.echo.DELETE("/api/:app", runner.deleteApp)
	runner.echo.GET("/api/health", runner.health)
	runner.echo.Any("/:app/*", runner.proxyRequest)
}

func (runner *GoRunnerWebServer) deleteApp(c echo.Context) error {
	appName := c.Param("app")
	app, err := runner.runner.GetApp(appName)
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

	runner.runner.DeleteApp(appName)

	return c.JSON(http.StatusOK, app)
}

func (runner *GoRunnerWebServer) appStatus(c echo.Context) error {
	appName := c.Param("app")
	runner.logger.Debug().Msgf("get app status: appName=%s", appName)

	goapp, err := runner.runner.GetApp(appName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	return c.JSON(http.StatusOK, goapp)
}

func (runner *GoRunnerWebServer) registerApp(c echo.Context) error {
	runner.logger.Info().Msg("new app...")
	params := new(DeployAppParams)
	err := c.Bind(params)
	if err != nil {
		runner.logger.Error().Msg("malformed request")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		runner.logger.Error().Msg("invalid request params")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	runner.logger.Info().Msgf("registering app... - app=%s, gitUrl=%s", params.App, params.GitUrl)
	goapp, err := runner.runner.GetApp(params.App)
	if goapp != nil {
		return c.JSON(http.StatusBadRequest, errStatus{
			goapp, errors.New("app already exists"),
		})
	}

	goapp, err = runner.runner.RegisterApp(params.App, params.GitUrl)
	if err != nil {
		runner.logger.Error().Msgf("error registering app. - app=%s, gitUrl=%s, err=%q", params.App, params.GitUrl, err)
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	return runner.deployApp(c, goapp)
}

func (runner *GoRunnerWebServer) updateApp(c echo.Context) error {
	runner.logger.Info().Msg("updating app...")
	params := new(UpdateAppParams)
	err := c.Bind(params)
	if err != nil {
		runner.logger.Error().Msg("malformed request")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		runner.logger.Error().Msg("invalid request params")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	app, err := runner.runner.GetApp(params.App)
	if err != nil {
		runner.logger.Error().Msgf("app not found. app=%s", params.App)
		return c.JSON(http.StatusNotFound, errStatus{
			nil, err,
		})
	}

	switch params.Action {
	case "deploy":
		return runner.deployApp(c, app)
	}

	return c.String(http.StatusInternalServerError, "Unknown error")
}
