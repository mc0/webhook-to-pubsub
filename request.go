package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"log"
	"net/url"
	"time"
)

const (
	// Fulfillment is a basic request
	Fulfillment = "fulfill"
	// ErrorGettingPayload is used in error strings
	ErrorGettingPayload = "error getting payload"
)

// Request is an instance of an incoming http request that has been processed
// into the constitute parts.
type Request struct {
	*RequestPayload
	Time int64
	Type string
	UUID string
}

// RequestPayload are the fundamental pieces of the request with Value being
// extracted from the payload.
type RequestPayload struct {
	RemoteAddr string
	UserAgent  string
	Value      string
}

// NewRequest takes the payload data and creates a proper request
// with a unique identifier and time filled into a `Request`.
func NewRequest(data *RequestPayload, eventType string) *Request {
	now := time.Now()
	u, _ := uuid.NewV4()
	event := &Request{
		RequestPayload: data,
		Time:           now.Unix(),
		Type:           eventType,
		UUID:           u.String(),
	}
	return event
}

func getPayloadFromQuery(query url.Values, event string) (payload *RequestPayload, err error) {
	val, ok := query["p"]
	if !ok {
		return nil, errors.New(ErrorGettingPayload)
	}

	decrypted, err := decryptURLBase64(encryptBlock, val[0])
	if nil != err {
		log.Print("decrypt failed: ", err)
		return nil, err
	}

	payload = &RequestPayload{}
	switch event {
	case Fulfillment:
		payload.Value = string(decrypted)
	default:
		err := fmt.Errorf("Unknown event: %s", event)
		log.Println(err)
		return nil, err
	}

	return payload, err
}

func generateRequestPayload(link string) (result string, err error) {
	data, err := encrypt(encryptBlock, []byte(link))
	if nil != err {
		return "", err
	}

	result = base64.URLEncoding.EncodeToString(data)

	return result, err
}
