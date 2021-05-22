package web

import (
	"fmt"
	"github.com/JackKCWong/go-runner/internal/app"
	"github.com/labstack/echo/v4"
	"net/http"
)

type GoRunnerWebAPI struct {
	server *echo.Echo
	addr   string
	runner *app.GoRunner
}

type errStatus struct {
	*app.GoApp
	Error error
}

func NewWebServer(wd, addr string) *GoRunnerWebAPI {
	return &GoRunnerWebAPI{
		server: echo.New(),
		addr:   addr,
		runner: app.NewGoRunner(wd),
	}
}

func (s *GoRunnerWebAPI) Start() error {
	s.server.GET("/api/:app", s.appStatus)
	s.server.POST("/api/:app", s.deployApp)
	s.server.DELETE("/api/:app", s.deleteApp)

	s.server.Any("/:app", s.proxyRequest)

	return s.server.Start(s.addr)
}

func (s *GoRunnerWebAPI) deleteApp(c echo.Context) error {
	appName := c.Param("app")
	app, err := s.runner.GetApp(appName)
	if err != nil {
		return err
	}

	app.Stop()

	return c.JSON(http.StatusOK, app)
}

func (s *GoRunnerWebAPI) appStatus(c echo.Context) error {
	appName := c.Param("app")
	goapp, err := s.runner.GetApp(appName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	return c.JSON(http.StatusOK, goapp)
}

func (s *GoRunnerWebAPI) deployApp(c echo.Context) error {
	appName := c.Param("app")
	gitUrl := c.Param("gitUrl")
	goapp, err := s.runner.RegisterApp(appName, gitUrl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	err = s.runner.StartApp(appName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	return c.JSON(http.StatusOK, goapp)
}

func (s *GoRunnerWebAPI) proxyRequest(c echo.Context) error {
	appName := c.Param("app")
	goapp, err := s.runner.GetApp(appName)
	if err != nil {
		return err
	}

	resp, err := goapp.Handle(c.Request())
	if err != nil {
		return err
	}

	err = resp.Write(c.Response().Writer)
	if err != nil {
		return err
	}

	return nil
}
