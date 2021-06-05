package web

import (
	"context"
	"fmt"
	"github.com/JackKCWong/go-runner/internal/core"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"io"
	"net/http"
	"sync"
)

type GoRunnerWebServer struct {
	_ struct{}
	sync.Mutex
	server     *echo.Echo
	runner     *core.GoRunner
	Status     string `json:"status"`
	WorkingDir string `json:"working_dir"`
}

type errStatus struct {
	*core.GoApp
	Error error
}

func NewWebServer(wd string) *GoRunnerWebServer {
	return &GoRunnerWebServer{
		server:     echo.New(),
		Status:     "NEW",
		WorkingDir: wd,
		runner:     core.NewGoRunner(wd),
	}
}

func (s *GoRunnerWebServer) Start(addr string) error {
	s.Lock()

	s.server.Logger.SetLevel(log.DEBUG)
	err := s.runner.Rehydrate()
	if err != nil {
		return err
	}

	s.server.GET("/api/:app", s.appStatus)
	s.server.POST("/api/:app", s.deployApp)
	s.server.DELETE("/api/:app", s.deleteApp)
	s.server.GET("/api/health", s.health)
	s.server.Any("/:app/*", s.proxyRequest)

	s.Status = "STARTED"
	s.Unlock()

	return s.server.Start(addr)
}

func (s *GoRunnerWebServer) Stop(c context.Context) error {
	s.Lock()
	defer s.Unlock()

	_ = s.server.Shutdown(c)
	_ = s.runner.Stop(c)

	return nil
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
	goapp, err := s.runner.GetApp(appName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	return c.JSON(http.StatusOK, goapp)
}

func (s *GoRunnerWebServer) deployApp(c echo.Context) error {
	params := new(DeployAppParams)
	err := c.Bind(params)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errStatus{
			nil, err,
		})
	}

	goapp, err := s.runner.RegisterApp(params.App, params.GitUrl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	err = s.runner.StartApp(goapp.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	return c.JSON(http.StatusOK, goapp)
}

func (s *GoRunnerWebServer) proxyRequest(c echo.Context) error {
	appName := c.Param("app")
	goapp, err := s.runner.GetApp(appName)
	if err != nil {
		return err
	}

	resp, err := goapp.Handle(c.Request())
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	c.Response().WriteHeader(resp.StatusCode)

	for k, vals := range resp.Header {
		for _, v := range vals {
			c.Response().Header().Add(k, v)
		}
	}

	_, err = io.Copy(c.Response().Writer, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (s *GoRunnerWebServer) health(c echo.Context) error {
	return c.JSON(http.StatusOK, s)
}
