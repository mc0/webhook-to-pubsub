package main

import (
	"testing"
)

func TestRequest(t *testing.T) {
	payload := &RequestPayload{false, "192.168.0.1", "Mac", "http://google.com"}

	r := NewRequest(payload, Fulfillment)
	if r.Time != 0 {
		t.Fatal("NewRequest returned invalid data for req Time", r)
	}
	if r.Type != "fulfillment" {
		t.Fatal("NewRequest returned invalid data for req Type", r)
	}
	if r.Value != "http://google.com" {
		t.Fatal("NewRequest returned invalid data for req Value", r)
	}
	if r.RemoteAddr != "192.168.0.1" {
		t.Fatal("NewRequest returned invalid data for req RemoteAddr", r)
	}
	if r.UserAgent != "Mac" {
		t.Fatal("NewRequest returned invalid data for req UserAgent", r)
	}
}
