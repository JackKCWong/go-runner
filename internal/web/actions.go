package web

import (
	"github.com/JackKCWong/go-runner/internal/core"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (server *GoRunnerWebServer) deployApp(c echo.Context, goapp *core.GoApp) error {
	server.logger.Info().Msgf("deploying app... - app=%s, gitUrl=%s", goapp.Name, goapp.GitURL)
	err := goapp.Rebuild()
	if err != nil {
		server.logger.Error().Err(err).Msgf("failed to build. app=%s", goapp.Name)
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	err = goapp.Start()
	if err != nil {
		server.logger.Error().Err(err).Msgf("failed to start. app=%s", goapp.Name)
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	server.logger.Info().Msgf("app started. app=%s", goapp.Name)

	return c.JSON(http.StatusOK, goapp)
}

func (server *GoRunnerWebServer) restartApp(c echo.Context, goapp *core.GoApp) error {
	server.logger.Info().Msgf("restarting app... - app=%s, gitUrl=%s", goapp.Name, goapp.GitURL)
	err := goapp.Stop()
	if err != nil {
		server.logger.Error().Err(err).Msgf("failed to stop. app=%s", goapp.Name)
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	err = goapp.Start()
	if err != nil {
		server.logger.Error().Err(err).Msgf("failed to start. app=%s", goapp.Name)
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	server.logger.Info().Msgf("app restarted. - app=%s", goapp.Name)

	return c.JSON(http.StatusOK, goapp)
}
