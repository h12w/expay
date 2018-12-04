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
	"log"
)

type config struct {
	Host    string
	Storage string
}

func main() {
	server, err := new()
	if err != nil {
		log.Fatal(err)
	}
	if err := server.run(); err != nil {
		log.Fatal(err)
	}
}
