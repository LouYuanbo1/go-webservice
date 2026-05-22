package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func GinMetricsMiddleware(mw *MetricsMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		path := c.Request.URL.Path
		status := c.Writer.Status()

		mw.Record(path, status, duration)
	}
}

func TestGinMetricsMiddleware(t *testing.T) {
	reg := prometheus.NewRegistry()

	mw, err := NewMetricsMiddleware(MetricsConfig{
		Namespace: "ginadapter",
	}, reg)
	assert.NoError(t, err, "NewMetricsMiddleware should not return an error")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/gin/adapter", GinMetricsMiddleware(mw), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
	})

	req := httptest.NewRequest("GET", "/gin/adapter", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "StatusCode should be 200")
	assert.Contains(t, w.Body.String(), "message", "Response body should contain message field")
}

func TestResponseRecorder(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := NewResponseRecorder(w)

	assert.Equal(t, http.StatusOK, recorder.StatusCode(), "StatusCode should be 200")

	// Test WriteHeader
	recorder.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, recorder.StatusCode(), "StatusCode should be 404")

	// Test WriteHeader with 5xx status code
	recorder.WriteHeader(http.StatusInternalServerError)
	assert.Equal(t, http.StatusInternalServerError, recorder.StatusCode(), "StatusCode should be 500")
}

func TestNewMetricsMiddleware(t *testing.T) {
	reg := prometheus.NewRegistry()

	mw, err := NewMetricsMiddleware(MetricsConfig{
		Namespace: "myapp",
		Subsystem: "api",
	}, reg)

	assert.NoError(t, err, "NewMetricsMiddleware should not return an error")
	// Test if the metrics are registered
	assert.NotNil(t, mw, "NewMetricsMiddleware should return a non-nil value")
}

func TestMetricsMiddleware_CustomCallback(t *testing.T) {
	reg := prometheus.NewRegistry()

	customCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "custom",
			Name:      "requests_by_method",
			Help:      "Requests by HTTP method",
		},
		[]string{"method"},
	)

	customHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "custom",
			Name:      "latency_by_status",
			Help:      "Latency by status code",
		},
		[]string{"status_class"},
	)

	mw, err := NewMetricsMiddleware(MetricsConfig{
		CustomCollectors: []prometheus.Collector{customCounter, customHistogram},
	}, reg)
	assert.NoError(t, err, "NewMetricsMiddleware should not return an error")

	mw.AddCustomCallback(func(m *RequestMetrics) {
		customCounter.WithLabelValues("GET").Inc()

		statusClass := "2xx"
		if m.StatusCode >= 500 {
			statusClass = "5xx"
		} else if m.StatusCode >= 400 {
			statusClass = "4xx"
		} else if m.StatusCode >= 300 {
			statusClass = "3xx"
		}
		customHistogram.WithLabelValues(statusClass).Observe(m.Duration)
	})

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := mw.Handler(nextHandler)

	req := httptest.NewRequest("GET", "/custom", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "StatusCode should be 200")
}

func TestMetricsMiddleware_OnlyCustom(t *testing.T) {
	reg := prometheus.NewRegistry()

	customGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "active_requests",
			Help:      "Active requests",
		},
		[]string{"path"},
	)

	mw, err := NewMetricsMiddleware(MetricsConfig{
		CustomCollectors: []prometheus.Collector{customGauge},
	}, reg)
	assert.NoError(t, err)

	mw.AddCustomCallback(func(m *RequestMetrics) {
		customGauge.WithLabelValues(m.Path).Set(m.Duration * 1000)
	})

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	wrapped := mw.Handler(nextHandler)

	req := httptest.NewRequest("POST", "/api/create", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "StatusCode should be 201")
}

func TestMetricsMiddleware_Gin(t *testing.T) {
	reg := prometheus.NewRegistry()

	mw, err := NewMetricsMiddleware(MetricsConfig{
		Namespace: "ginapp",
	}, reg)
	assert.NoError(t, err, "NewMetricsMiddleware should not return an error")

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"status": "ok"})
	})

	wrapped := mw.Handler(r)

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "StatusCode should be 201")
	assert.Contains(t, w.Body.String(), "status", "Response body should contain status field")
}

func TestMetricsMiddleware_ErrorStatus(t *testing.T) {
	reg := prometheus.NewRegistry()

	mw, err := NewMetricsMiddleware(MetricsConfig{
		Namespace: "test",
	}, reg)
	assert.NoError(t, err, "NewMetricsMiddleware should not return an error")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	wrapped := mw.Handler(nextHandler)

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code, "StatusCode should be 500")
}
