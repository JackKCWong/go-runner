package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/JackKCWong/go-runner/internal/web"
	"os"
	"os/signal"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	wd := flag.String("wd", cwd, "workding directory")
	addr := flag.String("addr", ":8080", "local address to listen on. default to :8080")

	flag.Parse()

	runner := web.NewGoRunnerServer(*wd)

	{
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, os.Interrupt, os.Kill)
		go func() {
			fmt.Println("press Ctrl+C to exit.")
			<-sigchan
			fmt.Println("Ctrl+C pressed.")
			close(sigchan)
			runner.Stop(context.Background())
			os.Exit(0)
		}()
	}

	err = runner.Start(*addr)
	if err != nil {
		panic(err)
		return
	}
}
