package web

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

func (runner *GoRunnerWebServer) proxyRequest(c echo.Context) error {
	runner.logger.Debug().Msg("proxying request...")
	appName := c.Param("app")
	goapp, err := runner.runner.GetApp(appName)
	if err != nil {
		runner.logger.Debug().Msgf("app not found. - app=%s", appName)
		return c.String(http.StatusNotFound, fmt.Sprintf("%q", err))
	}

	request := c.Request()
	runner.logger.Info().Msgf("handling request... - url=%s", request.URL)
	resp, err := goapp.Handle(request)
	if err != nil {
		runner.logger.Error().Msgf("failed to handle request. - url=%s, err=%q", request.URL, err)
		return err
	}

	defer resp.Body.Close()

	runner.logger.Debug().Msgf("writing status code... - %d", resp.StatusCode)
	c.Response().WriteHeader(resp.StatusCode)

	for k, vals := range resp.Header {
		for _, v := range vals {
			runner.logger.Debug().Msgf("writing header... - %s: %s", k, v)
			c.Response().Header().Add(k, v)
		}
	}

	runner.logger.Debug().Msg("writing body...")
	_, err = io.Copy(c.Response().Writer, resp.Body)
	if err != nil {
		return err
	}

	runner.logger.Info().Msgf("request responded. - url=%s", request.URL)

	return nil
}
