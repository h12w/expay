ExPay: an exercise of a payment API
===================================

### Overview

ExPay is an exercise of a RESTful payment API.

The API provides fetch (GET), list (GET without id), create (POST), update (PUT)
and delete (DELETE) RESTful operations for payments. ([ref: HTTP Methods](https://www.restapitutorial.com/lessons/httpmethods.html))

### Install

```bash
go get h12.io/expay

cd $GOPATH/src/h12.io/expay
make test

go install h12.io/expay/cmd/expay
expay -h
# expay -host [host] -storage [storage]
```

### Code layout

```
expay/ all domain types and constants
    cmd/ contain all main packages of services
        expay/ expay service main package
    db/boltdb a boltdb implementation of expay.DB interface
    service/ contain logic of all services
        payment/ payment service logic
    testdata/  data for testing
```

### Storage

Multiple storage backends could be supported given the following interface:

```go
	DB interface {
		Create(v interface{}) (id string, err error)
		Get(id string, v interface{}) error
		Delete(id string) error
		Update(id string, v interface{}) error
		List() (Iter, error)
	}
	Iter interface {
		Next() bool
		Scan(v interface{}) (id string, err error)
		Close() error
	}
```

Two backends are currently supported:

* boltdb: ACID persistent KV store
* fakeDB: a memory based DB for unit testing

### API Document

* SwaggerHub: https://app.swaggerhub.com/apis/h12w/expay-api/1.0.0
* PDF API Documents (see below if in PDF)
