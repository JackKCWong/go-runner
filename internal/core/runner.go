package core

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

func NewGoRunner(cwd string) *GoRunner {
	return &GoRunner{
		cwd: cwd,
	}
}

type GoRunner struct {
	_    struct{}
	apps sync.Map
	cwd  string
}

const appsDir = "goapps"

func (r *GoRunner) RegisterApp(appName, gitUrl string) (*GoApp, error) {
	appDir := path.Join(r.cwd, appsDir, appName)

	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		//fmt.Printf("registering app in [%s]", appDir)
	} else {
		return nil, fmt.Errorf("app with the same name already exist in [%s] ", appDir)
	}

	app := &GoApp{
		Name:   appName,
		GitURL: gitUrl,
		AppDir: appDir,
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

func (r *GoRunner) StopApp(appName string) error {
	app, err := r.GetApp(appName)
	if err != nil {
		return err
	}

	return app.Stop()
}

func (r *GoRunner) DeleteApp(appName string) {
	r.apps.Delete(appName)
}

func (r *GoRunner) GetApp(appName string) (*GoApp, error) {
	app, ok := r.apps.Load(appName)
	if !ok {
		return nil, errors.New("app not found")
	}

	return app.(*GoApp), nil
}

//Rehydrate brings up all the apps already in the app dir
func (r *GoRunner) Rehydrate() error {
	appsDir := path.Join(r.cwd, appsDir)
	dirs, err := ioutil.ReadDir(appsDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(appsDir, 0770)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			appDir := path.Join(r.cwd, appsDir, dir.Name())
			app := &GoApp{
				Name:   dir.Name(),
				AppDir: appDir,
			}

			_ = app.Reattach()
			_ = app.Start()

			r.apps.Store(app.Name, app)
		}
	}

	return nil
}

func (r *GoRunner) ListApps() []*GoApp {
	apps := make([]*GoApp, 0)
	r.apps.Range(func(_, app interface{}) bool {
		apps = append(apps, app.(*GoApp))
		return true
	})

	return apps
}

func (r *GoRunner) Stop(c context.Context) error {
	r.apps.Range(func(key, value interface{}) bool {
		_ = value.(*GoApp).Stop()
		return true
	})

	return nil
}
