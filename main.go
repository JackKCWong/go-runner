package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/JackKCWong/go-runner/internal/web"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	var stopWg sync.WaitGroup

	{
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, os.Interrupt, os.Kill)
		stopWg.Add(1)
		go func() {
			defer stopWg.Done()
			fmt.Println("press Ctrl+C to exit.")
			<-sigchan
			fmt.Println("Ctrl+C pressed.")
			close(sigchan)
			fmt.Println("stopping go-runner")
			runner.Stop(context.Background())
			fmt.Println("go-runner stopped")
		}()
	}

	err = runner.Bootsrap(*addr)
	if err != nil {
		panic(err)
	}

	runner.Serve()
	stopWg.Wait()
}
