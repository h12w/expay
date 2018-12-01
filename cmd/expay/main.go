// Package main ExPay API
//
// ExPay API provides a RESTful payment API
//
// Version: 1.0.0
//
// swagger:meta
//go:generate env SWAGGER_GENERATE_EXTENSION=false swagger generate spec -o swagger.json
//go:generate docker run --rm -v $PWD/cmd/expay:/opt swagger2markup/swagger2markup convert -i /opt/swagger.json -f /opt/swagger
//go:generate asciidoctor-pdf swagger.adoc
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"h12.io/expay"
	"h12.io/expay/db/boltdb"
	"h12.io/expay/service/payment"
)

type config struct {
	Host    string
	Storage string
}

func main() {
	cfg := &config{}
	flag.StringVar(&cfg.Host, "host", ":"+strconv.Itoa(expay.DefaultPort), "host of the expay service")
	flag.StringVar(&cfg.Storage, "storage", "storage.bolt", "config of the storage")
	flag.Parse()
	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg *config) error {
	db, err := boltdb.New(cfg.Storage)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:           cfg.Host,
		Handler:        payment.NewService(db.Bucket("payment")),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	notifyStop(server.Shutdown)

	log.Printf("ExPay service listening on %s", cfg.Host)
	return server.ListenAndServe()
}

// notifyStop listens to process signal and calls stopFn when received
func notifyStop(stopFn func(context.Context) error) {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-stopChan
		log.Printf("got signal %v", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := stopFn(ctx); err != nil {
			log.Print(err)
		}
	}()
}
