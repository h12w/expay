package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"h12.io/expay"
	"h12.io/expay/testdata"
)

func TestServerRun(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.Args = []string{"expay", "-host", "127.0.0.1:0", "-storage", path.Join(dir, "storage.bolt")}
	server, err := new()
	if err != nil {
		t.Fatal(err)
	}
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.run()
	}()

	client := &http.Client{Timeout: time.Second}
	urlPrefix := "http://" + server.listener.Addr().String() + "/v1/payments"

	// create
	id := ""
	{
		req, err := http.NewRequest("POST", urlPrefix, strings.NewReader(testdata.Payment))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected status code %d got %d", http.StatusCreated, resp.StatusCode)
		}
		payResp := &expay.PaymentResponse{}
		if err := json.NewDecoder(resp.Body).Decode(payResp); err != nil {
			t.Fatal(err)
		}
		if len(payResp.Data) != 1 {
			t.Fatalf("expect 1 payment returned but got %d", len(payResp.Data))
		}
		id = payResp.Data[0].ID
		resp.Body.Close()
	}

	// fetch
	{
		req, err := http.NewRequest("GET", urlPrefix+"/"+id, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d got %d", http.StatusOK, resp.StatusCode)
		}
		payResp := &expay.PaymentResponse{}
		if err := json.NewDecoder(resp.Body).Decode(payResp); err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()

		expectedPayments := []expay.Payment{{ID: id}}
		if err := json.Unmarshal([]byte(testdata.Payment), &expectedPayments[0]); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(payResp.Data, expectedPayments) {
			t.Fatalf("expect \n%+v\n got \n%+v", expectedPayments, payResp.Data)
		}
	}

	server.stopChan <- syscall.SIGINT

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
	default:
	}
}
