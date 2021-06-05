package app

import (
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
	apps sync.Map
	cwd  string
}

func (r *GoRunner) RegisterApp(appName, gitUrl string) (*GoApp, error) {
	appDir := path.Join(r.cwd, "apps", appName)

	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		fmt.Printf("registering app in [%s]", appDir)
	} else {
		return nil, fmt.Errorf("app with the same name already exist in [%s] ", appDir)
	}

	app := &GoApp{
		Name:   appName,
		GitURL: gitUrl,
		AppDir: appDir,
	}

	if err := app.Deploy(gitUrl); err != nil {
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

//Rehydrate brings up all the apps already in the app dir
func (r *GoRunner) Rehydrate() error {
	appsDir := path.Join(r.cwd, "apps")
	dirs, err := ioutil.ReadDir(appsDir)
	if os.IsNotExist(err) {
		err := os.Mkdir(appsDir, 0770)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			appDir := path.Join(r.cwd, "apps", dir.Name())
			app := &GoApp{
				Name:   dir.Name(),
				AppDir: appDir,
			}

			//err = app.Reload()
			//if err != nil {
			//	fmt.Printf("error loading app [%#v]", app)
			//	continue
			//}

			err = app.Start()
			if err != nil {
				fmt.Printf("error starting app [%#v]", app)
				continue
			}

			r.apps.Store(app.Name, app)
		}
	}

	return nil
}
