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

func (server *GoRunnerWebServer) Bootsrap(addr string) error {
	server.logger.Info().Msg("starting go-server server...")
	server.Lock()
	defer server.Unlock()

	server.logger.Info().Msg("rehydrating apps...")
	err := server.runner.Rehydrate()
	if err != nil {
		server.logger.Error().Msgf("errror during rehydration - %q", err)
		return err
	}

	server.setRoutes()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		server.logger.Error().
			Err(err).
			Str("addr", addr).
			Msg("cannot start listener")

		return err
	}

	server.echo.Listener = listener
	server.status = "STARTED"
	server.logger.Info().Msg("go server stared.")

	return nil
}

func (server *GoRunnerWebServer) Serve() error {
	return server.echo.Start("")
}

func (server *GoRunnerWebServer) Stop(c context.Context) error {
	server.Lock()
	defer server.Unlock()

	_ = server.echo.Shutdown(c)
	_ = server.runner.Stop(c)

	return nil
}

func (server *GoRunnerWebServer) health(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, server, "  ")
}

func (server *GoRunnerWebServer) MarshalJSON() ([]byte, error) {
	apps := server.runner.ListApps()
	addr := server.echo.ListenerAddr()
	return json.Marshal(struct {
		Status string        `json:"status"`
		Wd     string        `json:"workding_dir"`
		Addr   net.Addr      `json:"addr"`
		Apps   []*core.GoApp `json:"apps,omitempty"`
		NoApps int           `json:"no_of_apps"`
	}{
		server.status,
		server.wd,
		addr,
		apps,
		len(apps),
	})
}

func (server *GoRunnerWebServer) port() uint16 {
	server.Lock()
	defer server.Unlock()

	if server.status != "STARTED" {
		panic("server not started yet")
	}

	addr := server.echo.ListenerAddr()
	addrStr := addr.String()
	parts := strings.Split(addrStr, ":")
	port := parts[len(parts)-1]

	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		panic(fmt.Errorf("failed to parse listener address [%s], sth must be very wrong", addrStr))
	}

	return uint16(p)
}
