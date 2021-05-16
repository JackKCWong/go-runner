package main

import (
	"github.com/JackKCWong/go-runner/internal/app"
	"io/ioutil"
	"os"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	tmp, err := ioutil.TempFile(wd, "go-runner")
	if err != nil {
		return
	}

	runner := app.NewGoRunner(tmp.Name(), ":8080")

	err = runner.Start()
	if err != nil {
		panic(err)
		return
	}
}
