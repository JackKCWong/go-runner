package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	unixsock := flag.String("sock", "", "path to unix socket")

	flag.Parse()

	if len(*unixsock) == 0 {
		fmt.Println("missing required params")
		fmt.Printf("usage: %s -sock /path/to/socket\n", os.Args[0])
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

	// remove socket when exit
	{
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, os.Interrupt, os.Kill)
		go func() {
			fmt.Println("press Ctrl+C to exit.")
			<-sigchan
			fmt.Println("Ctrl+C pressed.")
			close(sigchan)
			os.Remove(*unixsock)
			os.Exit(0)
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/greeting", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("hello world"))
	})

	server := http.Server{
		Handler: mux,
	}

	err = server.Serve(listener)
}
