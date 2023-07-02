package tracer

import (
	"contrib.go.opencensus.io/exporter/zipkin"
	"fmt"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
	"log"
)

const (
	grpcServerPort   = ":50051"
	zipkinServerPort = "9411"
)

func NewExporter() *zipkin.Exporter {
	// 1. Configure exporter to export traces to Zipkin.
	localEndpoint, err := openzipkin.NewEndpoint("ecommerce service tracing", "localhost"+grpcServerPort)
	if err != nil {
		log.Fatalf("Failed to create the local zipkinEndpoint: %v", err)
	}
	reporter := zipkinHTTP.NewReporter(fmt.Sprintf("http://localhost:%s/api/v2/spans", zipkinServerPort))
	zipkinExporter := zipkin.NewExporter(reporter, localEndpoint)
	return zipkinExporter
}

func RegisterExporterWithTracer() {
	trace.RegisterExporter(NewExporter())
	// 2. Configure 100% sample rate, otherwise, few traces will be sampled.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
}
