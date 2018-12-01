package payment

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"h12.io/expay"
	"h12.io/expay/service"
)

const urlPrefix = "/v1/payments"

// Service provides a payment RESTful service
type Service struct {
	http.Handler
	db expay.DB
}

// fetchParam is the parameter for fetchPayment (for doc only)
//
// swagger:parameters fetchPayment
type fetchParam struct {
	// ID is payment ID
	//
	// in:path
	ID string `json:"id"`
}

// updateParam is the parameter for updatePayment (for doc only)
//
// swagger:parameters updatePayment
type updateParam struct {
	// ID is payment ID
	//
	// in:path
	ID string `json:"id"`
	// Payment info
	//
	// in:body
	Payment expay.Payment `json:"payment"`
}

// createParam is the parameter for createPayment (for doc only)
//
// swagger:parameters createPayment
type createParam struct {
	// Payment info
	//
	// in:body
	Payment expay.Payment `json:"payment"`
}

// deleteParam is the parameter for deletePayment (for doc only)
//
// swagger:parameters deletePayment
type deleteParam struct {
	// ID is payment ID
	//
	// in:path
	ID string `json:"id"`
}

// PaymentResponse is an envelope for a payment response
//
// swagger:response PaymentResponse
type paymentResponseWrapper struct {
	// in:body
	Resp expay.PaymentResponse
}

// NewService creates a new payment service
func NewService(db expay.DB) *Service {
	mux := mux.NewRouter()
	s := &Service{Handler: mux, db: db}

	mux.Use(service.CommonMiddleware)
	mux.NotFoundHandler = service.CommonMiddleware(http.HandlerFunc(s.notFound))
	mux.HandleFunc(urlPrefix+"/{id}", s.getPayment).Methods("GET")

	// swagger:route GET /v1/payments listPayment
	//
	// List payments
	//
	// This will show all available payments
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Responses:
	//       200: PaymentResponse
	//       500: ErrorResponse
	mux.HandleFunc(urlPrefix, s.listPayment).Methods("GET")

	// swagger:route PUT /v1/payments/{id} updatePayment
	//
	// Update payment
	//
	// This will update the payment with the ID
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Responses:
	//       200: PaymentResponse
	//       400: ErrorResponse
	//       500: ErrorResponse
	mux.HandleFunc(urlPrefix+"/{id}", s.updatePayment).Methods("PUT")

	// swagger:route DELETE /v1/payments/{id} deletePayment
	//
	// Delete payment
	//
	// This will delete the payment with the ID
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Responses:
	//       200: PaymentResponse
	//       500: ErrorResponse
	mux.HandleFunc(urlPrefix+"/{id}", s.deletePayment).Methods("DELETE")

	// swagger:route POST /v1/payments createPayment
	//
	// Create payment
	//
	// This will create a new payment
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Responses:
	//       201: PaymentResponse
	//       400: ErrorResponse
	//       500: ErrorResponse
	mux.HandleFunc(urlPrefix, s.createPayment).Methods("POST")

	return s
}

func (s *Service) notFound(w http.ResponseWriter, req *http.Request) {
	service.Error(w, "api not found", http.StatusNotFound)
}

func (s *Service) getPayment(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	pay := expay.Payment{}
	if err := s.db.Get(id, &pay); err != nil {
		if err == expay.ErrNotFound {
			service.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		service.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pay.ID = id
	_ = json.NewEncoder(w).Encode(&expay.PaymentResponse{Data: []expay.Payment{pay}})
}

func (s *Service) createPayment(w http.ResponseWriter, req *http.Request) {
	pay := expay.Payment{}
	if err := json.NewDecoder(req.Body).Decode(&pay); err != nil {
		service.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := pay.Verify(); err != nil {
		service.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := s.db.Create(pay)
	if err != nil {
		service.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", urlPrefix+"/"+id)
	w.WriteHeader(http.StatusCreated)
	pay.ID = id
	_ = json.NewEncoder(w).Encode(&expay.PaymentResponse{Data: []expay.Payment{pay}})
}

func (s *Service) updatePayment(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	if err := s.db.Get(id, &expay.Payment{}); err != nil {
		if err == expay.ErrNotFound {
			service.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		service.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pay := expay.Payment{}
	if err := json.NewDecoder(req.Body).Decode(&pay); err != nil {
		service.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pay.ID = id
	if err := pay.Verify(); err != nil {
		service.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.db.Update(id, pay); err != nil {
		service.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(&expay.PaymentResponse{Data: []expay.Payment{pay}})
}

func (s *Service) deletePayment(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	if err := s.db.Delete(id); err != nil {
		service.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(&expay.PaymentResponse{})
}

func (s *Service) listPayment(w http.ResponseWriter, req *http.Request) {
	iter, err := s.db.List()
	if err != nil {
		service.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payments := []expay.Payment{}
	for iter.Next() {
		payment := expay.Payment{}
		id, err := iter.Scan(&payment)
		if err != nil {
			service.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		payment.ID = id
		payments = append(payments, payment)
	}
	if err := iter.Close(); err != nil {
		service.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	paymentResponse := &expay.PaymentResponse{
		Data: payments,
		Links: &expay.Links{
			Self: "/v1/payments",
		},
	}
	_ = json.NewEncoder(w).Encode(paymentResponse)
}
