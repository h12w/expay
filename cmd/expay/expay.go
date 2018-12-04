package main

import (
	"context"
	"flag"
	"log"
	"net"
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

// server is the main server object of the program
type server struct {
	listener net.Listener
	server   *http.Server
	stopChan chan os.Signal
}

// new creates a new server object from configurations
func new() (*server, error) {
	cfg := &config{}
	flag.StringVar(&cfg.Host, "host", ":"+strconv.Itoa(expay.DefaultPort), "host of the expay service")
	flag.StringVar(&cfg.Storage, "storage", "storage.bolt", "config of the storage")
	flag.Parse()

	db, err := boltdb.New(cfg.Storage)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", cfg.Host)
	if err != nil {
		return nil, err
	}

	httpServer := &http.Server{
		Addr:           cfg.Host,
		Handler:        payment.NewService(db.Bucket("payment")),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	stopChan := make(chan os.Signal)
	notifyStop(stopChan, httpServer.Shutdown)

	log.Printf("ExPay service listening on %s", cfg.Host)
	return &server{
		listener: listener,
		server:   httpServer,
		stopChan: stopChan,
	}, nil
}

func (s *server) run() error {
	return s.server.Serve(s.listener)
}

// notifyStop listens to process signal and calls stopFn when received
func notifyStop(stopChan chan os.Signal, stopFn func(context.Context) error) {
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
