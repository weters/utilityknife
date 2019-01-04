package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/weters/utilityknife/service"
)

const readTimeout = 5 * time.Second
const writeTimeout = 10 * time.Second

func main() {
	addr := flag.String("addr", ":80", "address to listen on")
	dataDir := flag.String("dataDir", "/var/lib/utilityknife", "directory to store the key/value data")
	flag.Parse()

	svc := service.New(*dataDir)

	srv := &http.Server{
		Addr:         *addr,
		Handler:      handlers.CombinedLoggingHandler(os.Stdout, handlers.ProxyHeaders(svc)),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	go func() {
		log.Printf("Listening on %s", *addr)
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			panic(err)
		}
	}()

	sig := make(chan os.Signal, 0)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig

	log.Printf("Shutting down server...")
	srv.Close()
	log.Printf("Done.")
}
