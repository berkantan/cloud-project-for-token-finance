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

	sleepTime := 100 + rand.IntN(701)
	//new_generated_number := strconv.Itoa(generated_number)

	//value, _ := time.ParseDuration(new_generated_number)
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)

	// delay test
	// fmt.Fprintf(w, "{\"status\": \"Payment Processed successfully\", \"delay_ms\": %d}", sleepTime)

	errRateChance := rand.IntN(100)

	if errRateChance < 2 {
		status = "failed"
		w.WriteHeader(http.StatusGatewayTimeout)
		fmt.Fprintf(w, "{\"status\": \"Payment Gateway Timeout\", \"error_code\": 504}")
		paymentRequest.WithLabelValues("failed").Inc()
	} else if errRateChance < 12 {
		status = "declined"
		w.WriteHeader(http.StatusPaymentRequired)
		fmt.Fprintf(w, `{"status": "Payment declined", "reason": "insufficient_funds", "error_code": 402}`)
		paymentRequest.WithLabelValues("declined").Inc()
	} else {
		paymentRequest.WithLabelValues("success").Inc()

		fmt.Fprintf(w, `{"status": "Payment Processed successfully"}`)
	}

}

func main() {

	http.HandleFunc("/pay", paymentHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9093", nil)
}

func init() {
	prometheus.MustRegister(paymentRequest, paymentDuration)
}
