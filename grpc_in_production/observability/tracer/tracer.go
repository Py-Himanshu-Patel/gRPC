package tracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"io"

	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

func NewTracer(servicename string) (opentracing.Tracer, io.Closer, error) {
	// load config from environment variables
	cfg := config.Configuration{
		ServiceName: servicename,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		//Create the Jaeger exporter with the collector endpoint,
		//service name, and agent endpoint
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "127.0.0.1:6831",
		},
	}

	jLogger := log.StdLogger
	metricsFactory := prometheus.New()
	return cfg.NewTracer(
		config.Logger(jLogger),
		config.Metrics(metricsFactory),
	)
}
