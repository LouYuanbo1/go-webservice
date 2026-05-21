package monitor

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsConfig struct {
	Namespace        string
	Subsystem        string
	CustomCollectors []prometheus.Collector
}

type RequestMetrics struct {
	Path       string
	Status     string
	Duration   float64
	StatusCode int
}

type ResponseRecorder struct {
	http.ResponseWriter
	code int
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{w, http.StatusOK}
}

func (rr *ResponseRecorder) WriteHeader(code int) {
	rr.code = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *ResponseRecorder) StatusCode() int {
	return rr.code
}

type MetricsMiddleware struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	customCallbacks []func(*RequestMetrics)
}

func NewMetricsMiddleware(config MetricsConfig, reg prometheus.Registerer) (*MetricsMiddleware, error) {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	mw := &MetricsMiddleware{
		customCallbacks: make([]func(*RequestMetrics), 0),
	}

	if config.CustomCollectors == nil {
		reqsTotal := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_total",
				Help:      "Total HTTP requests",
			},
			[]string{"path", "status"},
		)

		reqDuration := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "Request duration in seconds",
			},
			[]string{"path"},
		)

		if err := reg.Register(reqsTotal); err != nil {
			return nil, err
		}
		if err := reg.Register(reqDuration); err != nil {
			return nil, err
		}

		mw.requestsTotal = reqsTotal
		mw.requestDuration = reqDuration
	}

	for _, collector := range config.CustomCollectors {
		if err := reg.Register(collector); err != nil {
			return nil, err
		}
	}

	return mw, nil
}

func (m *MetricsMiddleware) AddCustomCallback(callback func(*RequestMetrics)) {
	m.customCallbacks = append(m.customCallbacks, callback)
}

func (m *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rr := NewResponseRecorder(w)
		next.ServeHTTP(rr, r)

		dur := time.Since(start).Seconds()
		path := r.URL.Path
		statusCode := rr.StatusCode()
		status := strconv.Itoa(statusCode)

		if m.requestsTotal != nil {
			m.requestsTotal.WithLabelValues(path, status).Inc()
		}
		if m.requestDuration != nil {
			m.requestDuration.WithLabelValues(path).Observe(dur)
		}

		metrics := &RequestMetrics{
			Path:       path,
			Status:     status,
			Duration:   dur,
			StatusCode: statusCode,
		}

		for _, cb := range m.customCallbacks {
			cb(metrics)
		}
	})
}
