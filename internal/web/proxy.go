package web

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
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
	resp, err := goapp.Handle(request)
	if err != nil {
		server.logger.Error().Msgf("failed to handle request. - url=%s, err=%q", request.URL, err)
		return err
	}

	defer resp.Body.Close()

	server.logger.Debug().Msgf("writing status code... - %d", resp.StatusCode)
	c.Response().WriteHeader(resp.StatusCode)

	for k, vals := range resp.Header {
		for _, v := range vals {
			server.logger.Debug().Msgf("writing header... - %s: %s", k, v)
			c.Response().Header().Add(k, v)
		}
	}

	server.logger.Debug().Msg("writing body...")
	_, err = io.Copy(c.Response().Writer, resp.Body)
	if err != nil {
		return err
	}

	server.logger.Info().Msgf("request responded. - url=%s", request.URL)

	return nil
}
