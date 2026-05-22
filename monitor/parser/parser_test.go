package parser

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestParseFromReader(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{
			name: "valid gauge metric",
			input: `# HELP test_gauge A test gauge
# TYPE test_gauge gauge
test_gauge{instance="localhost:9090"} 42.5
`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "valid counter metric",
			input: `# HELP test_counter A test counter
# TYPE test_counter counter
test_counter{job="test"} 100
`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid_metric",
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseFromReader(bytes.NewReader([]byte(tc.input)))
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, result, tc.wantLen)
		})
	}
}

func TestParseFromBytes(t *testing.T) {
	input := []byte(`# HELP test_bytes A test metric
# TYPE test_bytes gauge
test_bytes{label="value"} 3.14
`)

	result, err := ParseFromBytes(input)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "test_bytes", result[0].Name)
}

func TestParseFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(`# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET"} 100
`))
	}))
	defer server.Close()

	result, err := ParseFromURL(server.URL)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "http_requests_total", result[0].Name)
}

func TestParseFromURL_Failure(t *testing.T) {
	t.Run("invalid URL", func(t *testing.T) {
		_, err := ParseFromURL("http://invalid-url-12345.invalid")
		assert.Error(t, err)
	})

	t.Run("non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		_, err := ParseFromURL(server.URL)
		assert.Error(t, err)
	})
}

func TestGatherAndParse(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "A test counter",
	})
	prometheus.MustRegister(counter)
	defer prometheus.Unregister(counter)

	counter.Inc()

	result, err := GatherAndParse()
	assert.NoError(t, err)

	var foundFamily *MetricFamily
	for _, family := range result {
		if family.Name == "test_counter" {
			foundFamily = family
			break
		}
	}
	assert.NotNil(t, foundFamily, "test_counter not found in gathered metrics")
	assert.Len(t, foundFamily.Metrics, 1)
	assert.Equal(t, 1.0, foundFamily.Metrics[0].Value)
}

func TestFilterByLabel(t *testing.T) {
	families := []*MetricFamily{
		{
			Name: "test_metric",
			Type: dto.MetricType_GAUGE,
			Metrics: []*Metric{
				{Labels: map[string]string{"env": "prod", "instance": "server1"}, Value: 10},
				{Labels: map[string]string{"env": "staging", "instance": "server2"}, Value: 20},
				{Labels: map[string]string{"env": "prod", "instance": "server3"}, Value: 30},
			},
		},
	}

	filtered := FilterByLabel(families, "env", "prod")
	assert.Len(t, filtered, 1)
	assert.Len(t, filtered[0].Metrics, 2)

	filtered = FilterByLabel(families, "env", "nonexistent")
	assert.Len(t, filtered, 0)
}

func TestFindFamilyByName(t *testing.T) {
	families := []*MetricFamily{
		{Name: "metric_a", Type: dto.MetricType_GAUGE},
		{Name: "metric_b", Type: dto.MetricType_COUNTER},
	}

	found := FindFamilyByName(families, "metric_b")
	assert.NotNil(t, found)
	assert.Equal(t, "metric_b", found.Name)

	notFound := FindFamilyByName(families, "metric_c")
	assert.Nil(t, notFound)
}

func TestMetricFamily_Sum(t *testing.T) {
	mf := &MetricFamily{
		Metrics: []*Metric{
			{Value: 10.0},
			{Value: 20.0},
			{Value: 30.0},
		},
	}

	assert.Equal(t, 60.0, mf.Sum())

	emptyMF := &MetricFamily{Metrics: []*Metric{}}
	assert.Equal(t, 0.0, emptyMF.Sum())
}

func TestMetricFamily_Avg(t *testing.T) {
	mf := &MetricFamily{
		Metrics: []*Metric{
			{Value: 10.0},
			{Value: 20.0},
			{Value: 30.0},
		},
	}

	assert.Equal(t, 20.0, mf.Avg())

	emptyMF := &MetricFamily{Metrics: []*Metric{}}
	assert.Equal(t, 0.0, emptyMF.Avg())
}

func TestMetricFamily_Min(t *testing.T) {
	mf := &MetricFamily{
		Metrics: []*Metric{
			{Value: 30.0},
			{Value: 10.0},
			{Value: 20.0},
		},
	}

	assert.Equal(t, 10.0, mf.Min())

	emptyMF := &MetricFamily{Metrics: []*Metric{}}
	assert.Equal(t, 0.0, emptyMF.Min())
}

func TestMetricFamily_Max(t *testing.T) {
	mf := &MetricFamily{
		Metrics: []*Metric{
			{Value: 10.0},
			{Value: 30.0},
			{Value: 20.0},
		},
	}

	assert.Equal(t, 30.0, mf.Max())

	emptyMF := &MetricFamily{Metrics: []*Metric{}}
	assert.Equal(t, 0.0, emptyMF.Max())
}

func TestMetricFamily_String(t *testing.T) {
	mf := &MetricFamily{
		Name: "test_string_metric",
		Help: "A test metric for String()",
		Type: dto.MetricType_GAUGE,
		Metrics: []*Metric{
			{Labels: map[string]string{"label1": "value1"}, Value: 42.5, Timestamp: 0},
		},
	}

	str := mf.String()
	assert.Contains(t, str, "# HELP test_string_metric A test metric for String()")
}

func TestRegisterCollectors(t *testing.T) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_register_counter",
		Help: "A test counter for registration",
	})

	err := RegisterCollectors(counter)
	assert.NoError(t, err)

	defer prometheus.Unregister(counter)

	err = RegisterCollectors(counter)
	assert.Error(t, err)
}
