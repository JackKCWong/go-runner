package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

func (server *GoRunnerWebServer) setRoutes() {
	// general api
	server.echo.POST("/api/apps", server.registerApp)
	server.echo.GET("/api/health", server.health)

	// per app api
	server.echo.GET("/api/:app", server.appStatus)
	server.echo.GET("/api/:app/stdout", server.appStdout)
	server.echo.GET("/api/:app/stderr", server.appStderr)
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

	app.Stop()
	err = app.Purge()
	server.runner.DeleteApp(appName)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			app, err,
		})
	}

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

	goapp, _ := server.runner.GetApp(params.App)
	if goapp == nil {
		server.logger.Info().Msgf("registering app... - app=%s, gitUrl=%s", params.App, params.GitUrl)
		goapp, err = server.runner.NewApp(params.App, params.GitUrl)
		if err != nil {
			server.logger.Err(err).Msgf("error registering app. - app=%s, gitUrl=%s", params.App, params.GitUrl)
			return c.JSON(http.StatusInternalServerError, errStatus{
				goapp, err,
			})
		}
	} else {
		server.logger.Info().Msgf("app already exist... - app=%s, gitUrl=%s", goapp.Name, goapp.GitURL)
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

func (server *GoRunnerWebServer) appStdout(c echo.Context) error {
	appName := c.Param("app")
	server.logger.Debug().Msgf("get app stdout - appName=%s", appName)

	goapp, err := server.runner.GetApp(appName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	buf := make(chan string, 1000)
	goapp.StdoutTo(buf)

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
	c.Response().WriteHeader(http.StatusOK)

	for line := range buf {
		fmt.Fprintf(c.Response().Writer, "%s\n", line)
		c.Response().Flush()
		if err != nil {
			goapp.UnsubscribeStdout(buf)
			return err
		}
	}

	return nil
}

func (server *GoRunnerWebServer) appStderr(c echo.Context) error {
	appName := c.Param("app")
	server.logger.Debug().Msgf("get app stderr - appName=%s", appName)

	goapp, err := server.runner.GetApp(appName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	buf := make(chan string, 1000)
	goapp.StderrTo(buf)

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
	c.Response().WriteHeader(http.StatusOK)

	for line := range buf {
		fmt.Fprintf(c.Response().Writer, "%s\n", line)
		c.Response().Flush()
		if err != nil {
			goapp.UnsubscribeStderr(buf)
			return err
		}
	}

	return nil
}
