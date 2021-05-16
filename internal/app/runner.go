package app

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"sync"
)

func NewGoRunner(cwd, address string) *GoRunner {
	return &GoRunner{
		cwd:     cwd,
		address: address,
		server:  echo.New(),
	}
}

type GoRunner struct {
	apps    sync.Map
	cwd     string
	server  *echo.Echo
	address string
}

func (r *GoRunner) registerApp(name, gitUrl string) (*GoApp, error) {

	appDir := fmt.Sprintf("%s/apps/%s", r.cwd, name)

	app := &GoApp{
		Name:   name,
		GitURL: gitUrl,
		dir:    appDir,
	}

	if err := app.Clone(); err != nil {
		return nil, err
	}

	r.apps.Store(name, app)

	return app, nil
}

func (r *GoRunner) startApp(appName string) error {

	app, err := r.getApp(appName)
	if err != nil {
		return err
	}

	return app.Start()
}

func (r *GoRunner) getApp(appName string) (*GoApp, error) {
	app, ok := r.apps.Load(appName)
	if !ok {
		return nil, errors.New("app not found")
	}

	return app.(*GoApp), nil
}
