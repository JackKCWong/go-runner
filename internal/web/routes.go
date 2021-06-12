package web

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (s *GoRunnerWebServer) setRoutes() {
	s.server.GET("/api/:app", s.appStatus)
	s.server.POST("/api/:app", s.registerApp)
	s.server.PUT("/api/:app", s.updateApp)
	s.server.DELETE("/api/:app", s.deleteApp)
	s.server.GET("/api/health", s.health)
	s.server.Any("/:app/*", s.proxyRequest)
}

func (s *GoRunnerWebServer) deleteApp(c echo.Context) error {
	appName := c.Param("app")
	app, err := s.runner.GetApp(appName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			app, err,
		})
	}

	err = app.Stop()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			app, err,
		})
	}

	err = s.runner.DeleteApp(appName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			app, err,
		})
	}

	return c.JSON(http.StatusOK, app)
}

func (s *GoRunnerWebServer) appStatus(c echo.Context) error {
	appName := c.Param("app")
	s.logger.Debugf("get app status: appName=%s", appName)

	goapp, err := s.runner.GetApp(appName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	return c.JSON(http.StatusOK, goapp)
}

func (s *GoRunnerWebServer) registerApp(c echo.Context) error {
	s.logger.Info("deploying app...")
	params := new(DeployAppParams)
	err := c.Bind(params)
	if err != nil {
		s.logger.Error("malformed request")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		s.logger.Error("invalid request params")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	s.logger.Infof("registering app... - app=%s, gitUrl=%s", params.App, params.GitUrl)
	goapp, err := s.runner.RegisterApp(params.App, params.GitUrl)
	if err != nil {
		s.logger.Errorf("error registering app. - app=%s, gitUrl=%s, err=%q", params.App, params.GitUrl, err)
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	return s.deployApp(c, goapp)
}

func (s *GoRunnerWebServer) updateApp(c echo.Context) error {
	s.logger.Info("updating app...")
	params := new(UpdateAppParams)
	err := c.Bind(params)
	if err != nil {
		s.logger.Error("malformed request")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		s.logger.Error("invalid request params")
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	app, err := s.runner.GetApp(params.App)
	if err != nil {
		s.logger.Errorf("app not found. app=%s", params.App)
		return c.JSON(http.StatusNotFound, errStatus{
			nil, err,
		})
	}

	switch params.Action {
	case "deploy":
		return s.deployApp(c, app)
	}

	return c.String(http.StatusInternalServerError, "Unknown error")
}
