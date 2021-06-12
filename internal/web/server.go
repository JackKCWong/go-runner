package web

import (
	"context"
	"encoding/json"
	"github.com/JackKCWong/go-runner/internal/core"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"sync"
)

type GoRunnerWebServer struct {
	_ struct{}
	sync.Mutex
	server *echo.Echo
	runner *core.GoRunner
	status string
	wd     string
	logger echo.Logger
}

func NewGoRunnerServer(wd string) *GoRunnerWebServer {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: nil,
		Format: "${time_rfc3339} http\t${status}\t${method} ${uri} " +
			"${latency_human}\t" +
			"${bytes_in}b ${bytes_out}b" +
			"\n",
		CustomTimeFormat: "",
		Output:           nil,
	}))
	logger := e.Logger.(*log.Logger)
	logger.SetHeader("${time_rfc3339} ${level}\t${line}\t")
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

func (s *GoRunnerWebServer) Stop(c context.Context) error {
	s.Lock()
	defer s.Unlock()

	_ = s.server.Shutdown(c)
	_ = s.runner.Stop(c)

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
