package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/JackKCWong/go-runner/internal/web"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.DebugLevel)
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

	err = runner.Bootsrap(*addr)
	if err != nil {
		panic(err)
		return
	}

	runner.Serve()
}
