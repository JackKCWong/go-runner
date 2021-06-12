package web

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

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
