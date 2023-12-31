package main

import (
	"aci-adr-go-base/service/bal"
	"aci-adr-go-base/service/dal"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	dal.InitMongo()
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter(os.Getenv("STAGE_NAME"))
	go serveMetrics()
	bal.Connect(meter)
}

func serveMetrics() {
	log.Printf("serving metrics at localhost:8080/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe("0.0.0.0:8080", nil) //nolint:gosec // Ignoring G114: Use of net/http serve function that has no support for setting timeouts.
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}
