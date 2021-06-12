package web

import (
	"context"
	"encoding/json"
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
	server *echo.Echo
	runner *core.GoRunner
	status string `json:"status"`
	wd     string `json:"working_dir"`
	logger echo.Logger
}

type errStatus struct {
	*core.GoApp
	Error error
}

func (e *errStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		*core.GoApp
		Error string
	}{e.GoApp, fmt.Sprintf("%q", e.Error)})
}

func NewGoRunnerServer(wd string) *GoRunnerWebServer {
	e := echo.New()
	logger := e.Logger.(*log.Logger)
	logger.SetHeader("${time_rfc3339} ${level} ${line}")
	logger.SetLevel(log.DEBUG)

	return &GoRunnerWebServer{
		server: e,
		status: "NEW",
		wd:     wd,
		runner: core.NewGoRunner(wd),
		logger: logger,
	}
}

func (s *GoRunnerWebServer) Start(addr string) error {
	s.logger.Info("starting go-runner server...")
	s.Lock()

	s.server.Logger.SetLevel(log.DEBUG)
	s.logger.Info("rehydrating apps...")
	err := s.runner.Rehydrate()
	if err != nil {
		s.logger.Errorf("errror during rehydration - %q", err)
		return err
	}

	s.setRoutes()

	s.status = "STARTED"
	s.logger.Info("go runner stared.")
	s.Unlock()

	return s.server.Start(addr)
}

func (s *GoRunnerWebServer) setRoutes() {
	s.server.GET("/api/:app", s.appStatus)
	s.server.POST("/api/:app", s.deployApp)
	s.server.DELETE("/api/:app", s.deleteApp)
	s.server.GET("/api/health", s.health)
	s.server.Any("/:app/*", s.proxyRequest)
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
	s.logger.Debugf("get app status: appName=%s", appName)

	goapp, err := s.runner.GetApp(appName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	return c.JSON(http.StatusOK, goapp)
}

func (s *GoRunnerWebServer) deployApp(c echo.Context) error {
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

	s.logger.Infof("starting app... - app=%s", goapp.Name)
	err = s.runner.StartApp(goapp.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	s.logger.Infof("app started. - app=%s", goapp.Name)
	return c.JSON(http.StatusOK, goapp)
}

func (s *GoRunnerWebServer) proxyRequest(c echo.Context) error {
	s.logger.Debug("proxying request...")
	appName := c.Param("app")
	goapp, err := s.runner.GetApp(appName)
	if err != nil {
		s.logger.Debugf("app not found. - app=%s", appName)
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	request := c.Request()
	s.logger.Infof("handling request... - url=%s", request.URL)
	resp, err := goapp.Handle(request)
	if err != nil {
		s.logger.Errorf("failed to handle request. - url=%s, err=%q", request.URL, err)
		return err
	}

	defer resp.Body.Close()

	s.logger.Debugf("writing status code... - %s", resp.StatusCode)
	c.Response().WriteHeader(resp.StatusCode)

	for k, vals := range resp.Header {
		for _, v := range vals {
			s.logger.Debugf("writing header... - %s: %s", k, v)
			c.Response().Header().Add(k, v)
		}
	}

	s.logger.Debug("writing body...")
	_, err = io.Copy(c.Response().Writer, resp.Body)
	if err != nil {
		return err
	}

	s.logger.Infof("request responded. - url=%s", request.URL)

	return nil
}

func (s *GoRunnerWebServer) health(c echo.Context) error {
	return c.JSON(http.StatusOK, s)
}

func (s *GoRunnerWebServer) MarshalJSON() ([]byte, error) {
	apps := s.runner.ListApps()
	return json.Marshal(struct {
		Status string        `json:"status"`
		Wd     string        `json:"workding_dir"`
		Apps   []*core.GoApp `json:"apps"`
		NoApps int           `json:"no_of_apps"`
	}{
		s.status,
		s.wd,
		apps,
		len(apps),
	})
}
