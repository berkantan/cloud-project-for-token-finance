package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var paymentRequest = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "request_status_counter",
	}, []string{"status"})

func paymentHandler(w http.ResponseWriter, r *http.Request) {

	sleep_time := 100 + rand.IntN(701)
	//new_generated_number := strconv.Itoa(generated_number)

	//value, _ := time.ParseDuration(new_generated_number)
	time.Sleep(time.Duration(sleep_time) * time.Millisecond)

	// delay test
	// fmt.Fprintf(w, "{\"status\": \"Payment Processed successfully\", \"delay_ms\": %d}", sleep_time)

	err_rate_chance := rand.IntN(100)

	if err_rate_chance < 15 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"status\": \"Payment Gateway Timeout\", \"error_code\": 500}")
		paymentRequest.WithLabelValues("failed").Inc()
		return
	}

	paymentRequest.WithLabelValues("success").Inc()

	fmt.Fprintf(w, `{"status": "Payment Processed successfully"}`)
}

func main() {

	http.HandleFunc("/bar", paymentHandler)
	http.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe(":9093", nil)
}

func init() {
	prometheus.MustRegister(paymentRequest)
}
