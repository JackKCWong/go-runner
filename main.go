package main

import (
	"flag"
	"github.com/JackKCWong/go-runner/internal/web"
	"os"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	wd := flag.String("wd", cwd, "workding directory")
	addr := flag.String("addr", ":8080", "local address to listen on. default to :8080")

	runner := web.NewGoRunnerServer(*wd)

	err = runner.Start(*addr)
	if err != nil {
		panic(err)
		return
	}
}
