package examples

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func main() {
	unixsock := flag.String("unixsock", "", "path to unix socket")

	flag.Parse()

	if len(*unixsock) == 0 {
		fmt.Println("missing required params")
		fmt.Printf("usage: %s -unixsock /path/to/socket\n", os.Args[0])
		os.Exit(0)
	}

	err := os.Remove(*unixsock)
	if err != nil {
		fmt.Printf("failed to remove in-use sock: %s, %q", *unixsock, err)
	}

	listener, err := net.Listen("unix", *unixsock)
	if err != nil {
		fmt.Printf("failed to listen on %s. err: %q\n", *unixsock, err)
		os.Exit(1)
	}


	mux := http.NewServeMux()
	mux.HandleFunc("/greeting", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("hello world"))
	})

	server := http.Server{
		Handler: mux,
	}

	var wg sync.WaitGroup
	// remove socket when exit
	{
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, os.Interrupt, os.Kill)
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println("press Ctrl+C to exit.")
			<-sigchan
			fmt.Println("Ctrl+C pressed.")
			close(sigchan)
			server.Shutdown(context.Background())
			os.Remove(*unixsock)
		}()
	}

	_ = server.Serve(listener)
	wg.Wait()
	fmt.Println("shutdown completed")
}
