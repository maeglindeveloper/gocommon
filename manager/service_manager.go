package manager

import (
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin/zipkin-go-opentracing"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServiceManagerInterface interface
type ServiceManagerInterface interface {
	Init(serviceName, debugAddr, gRPCUrl, zipkinURL string) bool
	Start() bool

	// Init the logger instance
	initLogger() bool

	// Init the tracer instance
	initTracer() bool

	// Init the metrics instance
	initMetrics() bool
}

// ServiceMetrics Structure defining metrics (To improve)
type ServiceMetrics struct {
	Ints     metrics.Counter
	Chars    metrics.Counter
	Duration metrics.Histogram
}

// Config Structure defining Config (To improve)
type Config struct {
	ServiceName string
	DebugAddr   string
	ZipkinURL   string
	GRPCAddr    string
}

// ServiceManager Struct
type ServiceManager struct {
	Config  Config
	Metrics ServiceMetrics
	Logger  log.Logger
	Tracer  stdopentracing.Tracer
}

// Init Logger
func (manager *ServiceManager) initLogger() bool {
	manager.Logger = log.NewLogfmtLogger(os.Stderr)
	manager.Logger = log.With(manager.Logger, "ts", log.DefaultTimestampUTC)
	manager.Logger = log.With(manager.Logger, "caller", log.DefaultCaller)

	return true
}

// Init tracer
func (manager *ServiceManager) initTracer() bool {

	// Check for zipkin URL
	if manager.Config.ZipkinURL == "" {
		manager.Tracer = stdopentracing.GlobalTracer()
		return true
	}

	// Instanciate Zipkin tracer
	manager.Logger.Log("tracer", "Zipkin", "type", "OpenTracing", "URL", manager.Config.ZipkinURL)
	collector, err := zipkinot.NewHTTPCollector(manager.Config.ZipkinURL)
	if err != nil {
		manager.Logger.Log("err", err)
		return false
	}

	defer collector.Close()
	recorder := zipkinot.NewRecorder(collector, false, "localhost:80", manager.Config.ServiceName)
	manager.Tracer, err = zipkinot.NewTracer(recorder)
	if err != nil {
		manager.Logger.Log("err", err)
		return false
	}

	return true
}

// Init metrics
func (manager *ServiceManager) initMetrics() bool {

	// Business-level metrics.
	manager.Metrics.Ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "example",
		Subsystem: manager.Config.ServiceName,
		Name:      "integers_summed",
		Help:      "Total count of integers summed via the Sum method.",
	}, []string{})

	manager.Metrics.Chars = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "example",
		Subsystem: manager.Config.ServiceName,
		Name:      "characters_concatenated",
		Help:      "Total count of characters concatenated via the Concat method.",
	}, []string{})

	// Endpoint-level metrics.
	manager.Metrics.Duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "example",
		Subsystem: manager.Config.ServiceName,
		Name:      "request_duration_seconds",
		Help:      "Request duration in seconds.",
	}, []string{"method", "success"})

	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	return true
}

// Init the service manager
func (manager *ServiceManager) Init(serviceName, debugAddr, gRPCAddr, zipkinURL string) bool {

	manager.Config.ServiceName = serviceName
	manager.Config.DebugAddr = debugAddr
	manager.Config.GRPCAddr = gRPCAddr
	manager.Config.ZipkinURL = zipkinURL

	// Init
	return (manager.initLogger() && manager.initTracer() && manager.initMetrics())
}

// Start the service manager
func (manager *ServiceManager) Start() bool {
	return true
}
