package web

import (
	"github.com/JackKCWong/go-runner/internal/core"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (s *GoRunnerWebServer) deployApp(c echo.Context, goapp *core.GoApp) error {
	s.logger.Infof("deploying app... - app=%s, gitUrl=%s", goapp.Name, goapp.GitURL)
	err := goapp.FetchAndBuild()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	err = goapp.Start()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errStatus{
			goapp, err,
		})
	}

	s.logger.Infof("app started. - app=%s", goapp.Name)

	return c.JSON(http.StatusOK, goapp)
}
