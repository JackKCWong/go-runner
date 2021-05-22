package app

import (
	"errors"
	"path"
	"sync"
)

func NewGoRunner(cwd string) *GoRunner {
	return &GoRunner{
		cwd: cwd,
	}
}

type GoRunner struct {
	apps sync.Map
	cwd  string
}

func (r *GoRunner) RegisterApp(appName, gitUrl string) (*GoApp, error) {

	appDir := path.Join(r.cwd, "apps", appName)

	app := &GoApp{
		Name:   appName,
		GitURL: gitUrl,
		AppDir: appDir,
	}

	if err := app.Clone(gitUrl); err != nil {
		return nil, err
	}

	r.apps.Store(appName, app)

	return app, nil
}

func (r *GoRunner) StartApp(appName string) error {

	app, err := r.GetApp(appName)
	if err != nil {
		return err
	}

	return app.Start()
}

func (r *GoRunner) GetApp(appName string) (*GoApp, error) {
	app, ok := r.apps.Load(appName)
	if !ok {
		return nil, errors.New("app not found")
	}

	return app.(*GoApp), nil
}
