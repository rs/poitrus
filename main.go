package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/rs/poitrus/overlay"
	"github.com/rs/poitrus/store"
)

func main() {
	listen := flag.String("listen", ":8080", "Address to listen on")
	origin := flag.String("overlay", "50.112.122.158", "IP of the service to overlay")
	root := flag.String("root", "/tmp/poitrus", "Path to the root directory")
	flag.Parse()

	h := http.NewServeMux()
	sh := store.Handler{Store: store.Store{Root: *root}}
	h.Handle("/", overlay.Handler(sh, *origin))
	s := &http.Server{
		Addr:           *listen,
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
