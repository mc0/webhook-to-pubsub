package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/tylerb/graceful.v1"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"syscall"
	"time"
)

var (
	httpPort        = new(string)
	encryptKey      = new(string)
	pubSubCredsJSON = new(string)
	debug           = flag.Bool("debug", false, "Debug mode")
	encryptBlock    cipher.Block
	jobChan         chan TopicMessage
	metrics         = struct {
		FulfillmentRequest         prometheus.Summary
		FulfillmentRequestFailures prometheus.Counter
		PubSubMessages             *prometheus.CounterVec
		PubSubFailures             *prometheus.CounterVec
	}{
		prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace: "web",
			Subsystem: "webhooks",
			Name:      "open_request",
			Help:      "Request latency for opens.",
		}),
		prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "web",
			Subsystem: "webhooks",
			Name:      "open_request_failures",
			Help:      "How many open requests have failed.",
		}),
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "web",
			Subsystem: "webhooks",
			Name:      "pubsub_messages",
			Help:      "Messages placed in the topic via a publish.",
		}, []string{"topic"}),
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "web",
			Subsystem: "webhooks",
			Name:      "pubsub_failures",
			Help:      "Failures when attempting to publish messages.",
		}, []string{"topic"}),
	}
)

func parseEnv(target *string, name, defaultValue string) {
	if v := os.Getenv(name); v != "" {
		*target = v
	} else {
		*target = defaultValue
	}
}

func init() {
	parseEnv(httpPort, "HTTP_PORT", "8080")
	flag.StringVar(httpPort, "httpPort", *httpPort, "Listen Address")
	parseEnv(encryptKey, "KEY", "")
	flag.StringVar(encryptKey, "key", *encryptKey, "AES Encryption key")
	parseEnv(pubSubCredsJSON, "PUBSUB_CREDS", "{}")
	flag.StringVar(pubSubCredsJSON, "pubSubCreds", *pubSubCredsJSON, "Json of pubsub specifiers of the form: {ProjectID, TopicName}")
	flag.Parse()

	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid:%d ", syscall.Getpid()))

	prometheus.MustRegister(
		metrics.FulfillmentRequest,
		metrics.FulfillmentRequestFailures,
		metrics.PubSubMessages,
		metrics.PubSubFailures,
	)
}

func main() {
	var err error
	encryptBlock, err = aes.NewCipher([]byte(*encryptKey))
	if err != nil {
		log.Fatal("AES block failed ", err)
	}

	var spec *PubSubSpec
	err = json.Unmarshal([]byte(*pubSubCredsJSON), &spec)
	if err != nil {
		log.Fatal("Invalid pubsub creds json: ", err)
	}

	jobChan = startPubSubChannel(spec)
	serveHTTP()
}

func serveHTTP() {
	mux := http.NewServeMux()

	// Handle anything /fulfillment/*
	mux.HandleFunc("/fulfillment/", handleFulfillment)

	mux.Handle("/metrics", promhttp.Handler())

	if *debug {
		mux.HandleFunc("/getKey", handleGetKey)
	}

	log.Println("serving on", *httpPort)
	graceful.Run(fmt.Sprintf(":%s", *httpPort), 15*time.Second, mux)
}

func handleRecover(w http.ResponseWriter, r *http.Request) {
	err := recover()
	if nil != err {
		w.WriteHeader(http.StatusInternalServerError)
		stack := make([]byte, 1<<16)
		runtime.Stack(stack, false)
		log.Printf("HTTP error caught %s\n%s", err, stack)
	}
}

func handleFulfillment(w http.ResponseWriter, r *http.Request) {
	var err error
	defer handleRecover(w, r)
	defer (func(start time.Time) {
		metrics.FulfillmentRequest.Observe(time.Since(start).Seconds())
		if err != nil {
			metrics.FulfillmentRequestFailures.Inc()
		}
	})(time.Now())
	query := r.URL.Query()
	payload, err := getPayloadFromQuery(query, Fulfillment)
	if nil != err {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	payload.RemoteAddr = getIPFromRequest(r)
	payload.UserAgent = r.UserAgent()

	open := NewRequest(payload, Fulfillment)
	msg, err := json.Marshal(open)
	if nil != err {
		log.Printf("error json.Marshal for open: %s, %s\n", err, open)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jobChan <- TopicMessage{
		Msg: string(msg),
	}
	w.WriteHeader(http.StatusOK)

	if *debug {
		fmt.Printf("Fulfillment: %q\n", payload)
	}
}

func handleGetKey(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var (
		link = ""
	)
	if val, ok := query["url"]; ok {
		link = val[0]
	}

	payload, err := generateRequestPayload(link)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to encrypt! %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\nencrypts to\n%v \n", link, url.QueryEscape(payload))
}
