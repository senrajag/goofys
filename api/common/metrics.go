package common

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const subsystemHTTPOutgoing = "http_outgoing"
const subsystemFuseOps = "fuse"
const namespace = "s3"

func InitMetricsHandler() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()
}

func GetMetricsHTTPTransport(transport http.RoundTripper) (http.RoundTripper, error) {
	i := &outgoingInstrumentation{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystemHTTPOutgoing,
				Name:      "requests_total",
				Help:      "A counter for outgoing requests from the wrapped client.",
			},
			[]string{"code", "method"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystemHTTPOutgoing,
				Name:      "request_duration_histogram_seconds",
				Help:      "A histogram of outgoing request latencies.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		firstByteDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystemHTTPOutgoing,
				Name:      "firstbyte_duration_histogram_seconds",
				Help:      "Time taken to get first byte latency histogram.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"event"},
		),
		inflight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemHTTPOutgoing,
			Name:      "in_flight_requests",
			Help:      "A gauge of in-flight outgoing requests for the wrapped client.",
		}),
	}

	trace := &promhttp.InstrumentTrace{
		GotFirstResponseByte: func(t float64) {
			i.firstByteDuration.WithLabelValues("got_first_response_byte").Observe(t)
		},
	}

	return promhttp.InstrumentRoundTripperInFlight(i.inflight,
		promhttp.InstrumentRoundTripperCounter(i.requests,
			promhttp.InstrumentRoundTripperTrace(trace,
				promhttp.InstrumentRoundTripperDuration(i.duration, transport),
			),
		),
	), prometheus.DefaultRegisterer.Register(i)
}

func ToPrometheusHttpClient(client *http.Client) (*http.Client, error) {
	return instrumentClient(client, prometheus.DefaultRegisterer)
}

func instrumentClient(c *http.Client, reg prometheus.Registerer) (*http.Client, error) {
	i := &outgoingInstrumentation{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystemHTTPOutgoing,
				Name:      "requests_total",
				Help:      "A counter for outgoing requests from the wrapped client.",
			},
			[]string{"code", "method"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystemHTTPOutgoing,
				Name:      "request_duration_histogram_seconds",
				Help:      "A histogram of outgoing request latencies.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		firstByteDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystemHTTPOutgoing,
				Name:      "firstbyte_duration_histogram_seconds",
				Help:      "Time taken to get first byte latency histogram.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"event"},
		),
		inflight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemHTTPOutgoing,
			Name:      "in_flight_requests",
			Help:      "A gauge of in-flight outgoing requests for the wrapped client.",
		}),
	}

	trace := &promhttp.InstrumentTrace{
		GotFirstResponseByte: func(t float64) {
			i.firstByteDuration.WithLabelValues("got_first_response_byte").Observe(t)
		},
	}

	transport := c.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	transport = promhttp.InstrumentRoundTripperInFlight(i.inflight,
		promhttp.InstrumentRoundTripperCounter(i.requests,
			promhttp.InstrumentRoundTripperTrace(trace,
				promhttp.InstrumentRoundTripperDuration(i.duration, transport),
			),
		),
	)
	return &http.Client{
		CheckRedirect: c.CheckRedirect,
		Jar:           c.Jar,
		Timeout:       c.Timeout,
		Transport: promhttp.InstrumentRoundTripperInFlight(i.inflight,
			promhttp.InstrumentRoundTripperCounter(i.requests,
				promhttp.InstrumentRoundTripperTrace(trace,
					promhttp.InstrumentRoundTripperDuration(i.duration, transport),
				),
			),
		),
	}, reg.Register(i)
}

type outgoingInstrumentation struct {
	duration          *prometheus.HistogramVec
	requests          *prometheus.CounterVec
	firstByteDuration *prometheus.HistogramVec
	inflight          prometheus.Gauge
}

// Describe implements prometheus.Collector interface.
func (i *outgoingInstrumentation) Describe(in chan<- *prometheus.Desc) {
	i.duration.Describe(in)
	i.requests.Describe(in)
	i.inflight.Describe(in)
	i.firstByteDuration.Describe(in)
}

// Collect implements prometheus.Collector interface.
func (i *outgoingInstrumentation) Collect(in chan<- prometheus.Metric) {
	i.duration.Collect(in)
	i.requests.Collect(in)
	i.inflight.Collect(in)
	i.firstByteDuration.Collect(in)
}

type MetricsClient struct {
	byteCounter *prometheus.CounterVec
	timeCounter *prometheus.CounterVec
}

func NewMetricsClient() *MetricsClient {
	return &MetricsClient{
		byteCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystemFuseOps,
				Name:      "bytes_total",
				Help:      "Bytes Read/Write to Fuse",
			},
			[]string{"operation"},
		),
		timeCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystemFuseOps,
				Name:      "time_taken",
				Help:      "Duration to Read/Write to Fuse",
			},
			[]string{"operation"},
		),
	}
}

func (m *MetricsClient) RecordFuseOps(c uint64, t int64, op string) {
	m.byteCounter.WithLabelValues(op).Add(float64(c))
	m.timeCounter.WithLabelValues(op).Add(float64(t))
}
