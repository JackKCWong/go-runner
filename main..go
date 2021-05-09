package main

import (
	"github.com/JackKCWong/go-runner/internal/rest"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"os"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	ioutil.TempFile(wd, "go-runner")
	e := echo.New()

	e.POST("/deploy", rest.DeployApp)
}
