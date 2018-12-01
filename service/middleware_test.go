package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCommonMiddleware(t *testing.T) {
	next := func(w http.ResponseWriter, req *http.Request) {}
	w := httptest.NewRecorder()
	CommonMiddleware(http.HandlerFunc(next)).ServeHTTP(w, &http.Request{})
	if value := w.Header().Get("Content-Type"); value != "application/json" {
		t.Fatalf("expect %s got %s", "application/json", value)
	}
}
