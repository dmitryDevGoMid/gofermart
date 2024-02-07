package jaeger

import (
	"context"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"

	config "github.com/dmitryDevGoMid/gofermart/internal/config"
	configyaml "github.com/dmitryDevGoMid/gofermart/internal/config/yaml"
)

// Init Jaeger
func InitJaeger(cfg *configyaml.Config) (opentracing.Tracer, io.Closer, error) {
	jaegerCfgInstance := jaegercfg.Configuration{
		ServiceName: cfg.Jaeger.ServiceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           cfg.Jaeger.LogSpans,
			LocalAgentHostPort: cfg.Jaeger.Host,
		},
	}

	return jaegerCfgInstance.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Metrics(metrics.NullFactory),
	)
}

//Создаем объект прослойку для трассировки запросо с условием конфига

type TraicingInterface interface {
	Tracing(ctx context.Context, label string) (opentracing.Span, context.Context)
}

type tracing struct {
	cfg *config.Config
}

func NewTracing(cfg *config.Config) TraicingInterface {
	return &tracing{
		cfg: cfg,
	}
}

func (t *tracing) Tracing(ctx context.Context, label string) (opentracing.Span, context.Context) {
	if !t.cfg.Server.TracingEnabled {
		return nil, ctx
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, label)
	return span, ctx
}
