package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	expectedCode := http.StatusBadRequest
	Error(w, "err msg", expectedCode)
	if w.Code != expectedCode {
		t.Fatalf("expect %d got %d", expectedCode, w.Code)
	}
	expectedBody := `{"code":400,"message":"err msg"}` + "\n"
	if body := w.Body.String(); body != expectedBody {
		t.Fatalf("expect %s got %s", expectedBody, body)
	}
}
