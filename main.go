package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/rs/poitrus/overlay"
	"github.com/rs/poitrus/shortner"
)

func main() {
	listen := flag.String("listen", ":8080", "Address to listen on")
	origin := flag.String("overlay", "50.112.122.158", "IP of the service to overlay")
	data := flag.String("data", "/tmp/poitrus", "Path to the data directory")
	flag.Parse()

	h := http.NewServeMux()
	db, err := shortner.NewDB(*data)
	if err != nil {
		log.Fatal(err)
	}
	sh := shortner.Handler{DB: db}
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
