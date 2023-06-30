package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpHandler(t *testing.T) {
	body := strings.NewReader(`
		{
		  "version": "4",
		  "groupKey": "test",
		  "truncatedAlerts": 0,
		  "status": "firing",
		  "receiver": "test",
		  "groupLabels": {},
		  "commonLabels": {},
		  "commonAnnotations": {},
		  "externalURL": "https://google.com",
		  "alerts": [
			{
			  "status": "firing",
			  "labels": {},
			  "annotations": {},
			  "startsAt": "2002-10-02T15:00:00Z",
			  "endsAt": "2002-10-02T15:00:00Z",
			  "generatorURL": "test",
			  "fingerprint": "test"
			}
		  ]
		}`)
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httpHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
}
