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

var paymentDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "payment_request_duration_seconds",
		Help:    "Ödeme İsteklerinin İşlenme Süresi",
		Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 1},
	}, []string{"status"})

func paymentHandler(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	status := "success"

	defer func() {
		paymentDuration.WithLabelValues(status).Observe(float64(time.Since(start).Seconds()))
	}()

	sleep_time := 100 + rand.IntN(701)
	//new_generated_number := strconv.Itoa(generated_number)

	//value, _ := time.ParseDuration(new_generated_number)
	time.Sleep(time.Duration(sleep_time) * time.Millisecond)

	// delay test
	// fmt.Fprintf(w, "{\"status\": \"Payment Processed successfully\", \"delay_ms\": %d}", sleep_time)

	err_rate_chance := rand.IntN(100)

	if err_rate_chance < 15 {
		status = "failed"
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"status\": \"Payment Gateway Timeout\", \"error_code\": 500}")
		paymentRequest.WithLabelValues("failed").Inc()
		fmt.Fprintf(w, `{"status": "Payment Processed successfully"}`)
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
	prometheus.MustRegister(paymentRequest, paymentDuration)
}
