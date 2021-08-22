package web

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (server *GoRunnerWebServer) proxyRequest(c echo.Context) error {
	server.logger.Debug().Msg("proxying request...")
	appName := c.Param("app")
	goapp, err := server.runner.GetApp(appName)
	if err != nil {
		server.logger.Debug().Msgf("app not found. - app=%s", appName)
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	if goapp.Status != "STARTED" {
		server.logger.Debug().Msgf("app not started. - app=%s, status=%s", goapp.Name, goapp.Status)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("app not started. - app=%s, status=%s", goapp.Name, goapp.Status))
	}

	request := c.Request()
	server.logger.Info().Msgf("handling request... - url=%s", request.URL)
	goapp.ServeHTTP(c.Response().Writer, request)
	server.logger.Debug().Msgf("request responded. - url=%s", request.URL)

	return nil
}
