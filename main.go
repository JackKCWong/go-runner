package main

import (
	"github.com/JackKCWong/go-runner/internal/web"
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

	runner := web.NewWebServer(tmp.Name())

	err = runner.Start(":8080")
	if err != nil {
		panic(err)
		return
	}
}
