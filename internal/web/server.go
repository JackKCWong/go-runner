package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JackKCWong/go-runner/internal/core"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v2"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type GoRunnerWebServer struct {
	_ struct{}
	sync.Mutex
	echo   *echo.Echo
	server *http.Server
	runner *core.GoRunner
	status string
	wd     string
	logger *zerolog.Logger
}

func NewGoRunnerServer(wd string) *GoRunnerWebServer {
	e := echo.New()
	e.HideBanner = true
	lechologger := lecho.From(log.Logger)
	e.Logger = lechologger
	e.Use(middleware.RequestID())
	e.Use(lecho.Middleware(lecho.Config{
		Logger:       lechologger,
		Skipper:      nil,
		RequestIDKey: "",
	}))

	return &GoRunnerWebServer{
		echo:   e,
		status: "NEW",
		wd:     wd,
		runner: core.NewGoRunner(wd),
		logger: &log.Logger,
	}
}

func (runner *GoRunnerWebServer) Bootsrap(addr string) error {
	runner.logger.Info().Msg("starting go-runner server...")
	runner.Lock()

	runner.logger.Info().Msg("rehydrating apps...")
	err := runner.runner.Rehydrate()
	if err != nil {
		runner.logger.Error().Msgf("errror during rehydration - %q", err)
		return err
	}

	runner.setRoutes()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		runner.logger.Error().
			Err(err).
			Str("addr", addr).
			Msg("cannot start listener")

		return err
	}

	runner.echo.Listener = listener
	runner.status = "STARTED"
	runner.logger.Info().Msg("go runner stared.")
	runner.Unlock()

	return nil
}

func (runner *GoRunnerWebServer) Serve() error {
	return runner.echo.Start("")
}

func (runner *GoRunnerWebServer) Stop(c context.Context) error {
	runner.Lock()
	defer runner.Unlock()

	_ = runner.echo.Shutdown(c)
	_ = runner.runner.Stop(c)

	return nil
}

func (runner *GoRunnerWebServer) health(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, runner, "  ")
}

func (runner *GoRunnerWebServer) MarshalJSON() ([]byte, error) {
	apps := runner.runner.ListApps()
	addr := runner.echo.ListenerAddr()
	return json.Marshal(struct {
		Status string        `json:"status"`
		Wd     string        `json:"workding_dir"`
		Addr   net.Addr      `json:"addr"`
		Apps   []*core.GoApp `json:"apps,omitempty"`
		NoApps int           `json:"no_of_apps"`
	}{
		runner.status,
		runner.wd,
		addr,
		apps,
		len(apps),
	})
}

func (runner *GoRunnerWebServer) port() uint16 {
	runner.Lock()
	defer runner.Unlock()

	if runner.status != "STARTED" {
		panic("server not started yet")
	}

	addr := runner.echo.ListenerAddr()
	addrStr := addr.String()
	parts := strings.Split(addrStr, ":")
	port := parts[len(parts)-1]

	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		panic(fmt.Errorf("failed to parse listener address [%s], sth must be very wrong", addrStr))
	}

	return uint16(p)
}
