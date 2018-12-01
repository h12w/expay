package payment

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"h12.io/expay"
	"h12.io/expay/testdata"
)

func TestPaymentService(t *testing.T) {
	t.Parallel()

	getReq := func(id string) func(string) *http.Request {
		return func(baseURL string) *http.Request {
			uri := baseURL + urlPrefix
			if id != "" {
				uri += "/" + id
			}
			req, _ := http.NewRequest(http.MethodGet, uri, nil)
			return req
		}
	}

	postReq := func(body string) func(string) *http.Request {
		return func(baseURL string) *http.Request {
			req, _ := http.NewRequest(http.MethodPost, baseURL+urlPrefix, strings.NewReader(body))
			return req
		}
	}

	putReq := func(id, body string) func(string) *http.Request {
		return func(baseURL string) *http.Request {
			req, _ := http.NewRequest(http.MethodPut, baseURL+urlPrefix+"/"+id, strings.NewReader(body))
			return req
		}
	}

	deleteReq := func(id string) func(string) *http.Request {
		return func(baseURL string) *http.Request {
			uri := baseURL + urlPrefix + "/" + id
			req, _ := http.NewRequest(http.MethodDelete, uri, nil)
			return req
		}
	}

	verifyCode := func(t *testing.T, resp *http.Response, expectedCode int) {
		t.Helper()
		if resp.StatusCode != expectedCode {
			t.Fatalf("expect HTTP status code %d (%s) but got %d (%s)",
				expectedCode, http.StatusText(expectedCode),
				resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		wantCt := "application/json"
		if ct := resp.Header.Get("Content-Type"); ct != wantCt {
			t.Fatalf("expect %s got %s", wantCt, ct)
		}
	}

	testcases := []struct {
		name string
		db   func() expay.DB
		req  func(baseURL string) *http.Request

		verify func(t *testing.T, resp *http.Response, s *Service)
	}{
		{
			name: "wrong url _ 404 not found",
			req: func(baseURL string) *http.Request {
				req, _ := http.NewRequest(http.MethodPost, baseURL+"/wrong_url", nil)
				return req
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusNotFound)
			},
		},

		{
			name: "fetch payment _ 404 not found",
			req:  getReq("nonexisted-id"),
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusNotFound)
			},
		},
		{
			name: "fetch payment _ 500 internal error",
			req:  getReq("id"),
			db: func() expay.DB {
				db := newFakeDB()
				db.getErr = errors.New("injected error")
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusInternalServerError)
			},
		},
		{
			name: "fetch payment _ 200 ok",
			req:  getReq("1"),
			db: func() expay.DB {
				db := newFakeDB()
				pay := &expay.Payment{}
				_ = json.Unmarshal([]byte(testdata.Payment), pay)
				db.m["1"] = pay
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusOK)
				respPay := &expay.PaymentResponse{}
				if err := json.NewDecoder(resp.Body).Decode(respPay); err != nil {
					t.Fatal(err)
				}
				if n := len(respPay.Data); n != 1 {
					t.Fatalf("expect 1 payment returned but got %d", n)
				}
				wantPay := expay.Payment{ID: "1"}
				_ = json.Unmarshal([]byte(testdata.Payment), &wantPay)
				if !reflect.DeepEqual(respPay.Data[0], wantPay) {
					t.Fatalf("expect %+v got %v", wantPay, respPay)
				}
			},
		},

		{
			name: "create payment _ 400 invalid request of empty body",
			req:  postReq(""),
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "create payment _ 400 missing field",
			req:  postReq("{}"),
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "create payment _ 500 internal error",
			req:  postReq(testdata.Payment),
			db: func() expay.DB {
				db := newFakeDB()
				db.createErr = errors.New("injected db error")
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusInternalServerError)
			},
		},
		{
			name: "create payment _ 201 created",
			req:  postReq(testdata.Payment),
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusCreated)

				inputPay := expay.Payment{}
				if err := json.Unmarshal([]byte(testdata.Payment), &inputPay); err != nil {
					t.Fatal(err)
				}

				respPay := &expay.PaymentResponse{}
				if err := json.NewDecoder(resp.Body).Decode(respPay); err != nil {
					t.Fatal(err)
				}
				if n := len(respPay.Data); n != 1 {
					t.Fatalf("expect 1 payment returned but got %d", n)
				}
				id := respPay.Data[0].ID
				if id == "" {
					t.Fatal("expect id but got empty string")
				}

				location := resp.Header.Get("Location")
				wantLocation := urlPrefix + "/" + id
				if location != wantLocation {
					t.Fatalf("expect location %s got %s", wantLocation, location)
				}

				inputPay.ID = id
				if !reflect.DeepEqual(respPay.Data[0], inputPay) {
					t.Fatalf("expect %+v got %+v", inputPay, respPay)
				}

				dbPay := expay.Payment{}
				if err := s.db.Get(id, &dbPay); err != nil {
					t.Fatal(err)
				}
				dbPay.ID = id
				if !reflect.DeepEqual(dbPay, respPay.Data[0]) {
					t.Fatalf("expect \n%+v\n got \n%+v", respPay, dbPay)
				}
			},
		},

		{
			name: "update payment _ 404 not found",
			req:  putReq("nonexisted-id", testdata.Payment),
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusNotFound)
			},
		},
		{
			name: "update payment _ 400 invalid request of empty body",
			db: func() expay.DB {
				db := newFakeDB()
				pay := &expay.Payment{ID: "1"}
				_ = json.Unmarshal([]byte(testdata.Payment), pay)
				db.m["1"] = pay
				return db
			},
			req: putReq("1", ""),
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "update payment _ 400 missing field",
			db: func() expay.DB {
				db := newFakeDB()
				pay := &expay.Payment{ID: "1"}
				_ = json.Unmarshal([]byte(testdata.Payment), pay)
				db.m["1"] = pay
				return db
			},
			req: putReq("1", "{}"),
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "update payment _ get error _ 500 internal error",
			req:  putReq("id", testdata.Payment),
			db: func() expay.DB {
				db := newFakeDB()
				db.getErr = errors.New("injected error")
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusInternalServerError)
			},
		},
		{
			name: "update payment _ update error _ 500 internal error",
			db: func() expay.DB {
				db := newFakeDB()
				pay := &expay.Payment{ID: "1"}
				_ = json.Unmarshal([]byte(testdata.Payment), pay)
				db.m["1"] = pay
				db.updateErr = errors.New("injected error")
				return db
			},
			req: putReq("1", testdata.Payment),
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusInternalServerError)
			},
		},
		{
			name: "update payment _ 200 ok",
			req:  putReq("1", testdata.Payment2),
			db: func() expay.DB {
				db := newFakeDB()
				pay := &expay.Payment{ID: "1"}
				_ = json.Unmarshal([]byte(testdata.Payment), pay)
				db.m["1"] = pay
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusOK)
				dbPay := &expay.Payment{}
				if err := s.db.Get("1", dbPay); err != nil {
					t.Fatal(err)
				}
				wantPay := &expay.Payment{ID: "1"}
				_ = json.Unmarshal([]byte(testdata.Payment2), wantPay)
				if !reflect.DeepEqual(dbPay, wantPay) {
					t.Fatalf("expect %+v got %v", wantPay, dbPay)
				}
			},
		},

		{
			name: "delete payment _ 500 internal error",
			req:  deleteReq("id"),
			db: func() expay.DB {
				db := newFakeDB()
				db.deleteErr = errors.New("injected error")
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusInternalServerError)
			},
		},
		{
			name: "delete payment _ 200 ok",
			req:  deleteReq("1"),
			db: func() expay.DB {
				db := newFakeDB()
				pay := &expay.Payment{}
				_ = json.Unmarshal([]byte(testdata.Payment), pay)
				db.m["1"] = pay
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusOK)

				if err := s.db.Get("1", &expay.Payment{}); err != expay.ErrNotFound {
					t.Fatalf("expect error %v got %v", expay.ErrNotFound, err)
				}
			},
		},

		{
			name: "list payments _ db error _ 500 internal error",
			req:  getReq(""),
			db: func() expay.DB {
				db := newFakeDB()
				db.m["1"] = &expay.Payment{}
				db.listErr = errors.New("injected error")
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusInternalServerError)
			},
		},
		{
			name: "list payments _ db iter scan error _ 500 internal error",
			req:  getReq(""),
			db: func() expay.DB {
				db := newFakeDB()
				db.m["1"] = &expay.Payment{}
				db.iterScanErr = errors.New("injected error")
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusInternalServerError)
			},
		},
		{
			name: "list payments _ db iter close error _ 500 internal error",
			req:  getReq(""),
			db: func() expay.DB {
				db := newFakeDB()
				db.m["1"] = &expay.Payment{}
				db.iterCloseErr = errors.New("injected error")
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusInternalServerError)
			},
		},
		{
			name: "list payments _ 200 ok",
			req:  getReq(""),
			db: func() expay.DB {
				db := newFakeDB()
				db.m["1"] = &expay.Payment{}
				db.m["2"] = &expay.Payment{}
				return db
			},
			verify: func(t *testing.T, resp *http.Response, s *Service) {
				verifyCode(t, resp, http.StatusOK)
				paymentResp := &expay.PaymentResponse{Data: []expay.Payment{}}
				if err := json.NewDecoder(resp.Body).Decode(paymentResp); err != nil {
					t.Fatal(err)
				}

				wantResp := &expay.PaymentResponse{
					Data: []expay.Payment{
						{ID: "1"}, {ID: "2"},
					},
					Links: &expay.Links{
						Self: "/v1/payments",
					},
				}
				if !reflect.DeepEqual(paymentResp, wantResp) {
					t.Fatalf("expect \n%+v\n got \n%+v", wantResp, paymentResp)
				}
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := expay.DB(newFakeDB())
			if tc.db != nil {
				db = tc.db()
			}
			paymentService := NewService(db)
			server := httptest.NewServer(paymentService)
			defer server.Close()

			resp, err := (&http.Client{}).Do(tc.req(server.URL))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if tc.verify != nil {
				tc.verify(t, resp, paymentService)
			}
		})
	}
}
