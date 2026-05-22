package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

type MetricFamily struct {
	Name    string
	Help    string
	Type    dto.MetricType
	Metrics []*Metric
}

type Metric struct {
	Labels    map[string]string
	Value     float64
	Timestamp int64
}

func ParseFromReader(reader io.Reader) ([]*MetricFamily, error) {
	parser := expfmt.NewTextParser(model.LegacyValidation)
	families, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", err)
	}

	return convertToMetricFamilies(families), nil
}

func ParseFromURL(url string) ([]*MetricFamily, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return ParseFromReader(resp.Body)
}

func ParseFromBytes(data []byte) ([]*MetricFamily, error) {
	return ParseFromReader(bytes.NewReader(data))
}

func GatherAndParse() ([]*MetricFamily, error) {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return nil, fmt.Errorf("failed to gather metrics: %w", err)
	}

	return convertToMetricFamiliesFromDTO(families), nil
}

func convertToMetricFamilies(families map[string]*dto.MetricFamily) []*MetricFamily {
	result := make([]*MetricFamily, 0, len(families))
	for name, family := range families {
		result = append(result, convertToMetricFamily(name, family))
	}
	return result
}

func convertToMetricFamiliesFromDTO(families []*dto.MetricFamily) []*MetricFamily {
	result := make([]*MetricFamily, 0, len(families))
	for _, family := range families {
		result = append(result, convertToMetricFamily(family.GetName(), family))
	}
	return result
}

func convertToMetricFamily(name string, family *dto.MetricFamily) *MetricFamily {
	mf := &MetricFamily{
		Name:    name,
		Help:    family.GetHelp(),
		Type:    family.GetType(),
		Metrics: make([]*Metric, 0, len(family.GetMetric())),
	}

	for _, m := range family.GetMetric() {
		metric := &Metric{
			Labels:    make(map[string]string),
			Value:     getMetricValue(m),
			Timestamp: m.GetTimestampMs(),
		}

		for _, label := range m.GetLabel() {
			metric.Labels[label.GetName()] = label.GetValue()
		}

		mf.Metrics = append(mf.Metrics, metric)
	}

	return mf
}

func getMetricValue(m *dto.Metric) float64 {
	switch {
	case m.Gauge != nil:
		return m.Gauge.GetValue()
	case m.Counter != nil:
		return m.Counter.GetValue()
	case m.Summary != nil:
		return m.Summary.GetSampleSum()
	case m.Histogram != nil:
		return m.Histogram.GetSampleSum()
	default:
		return 0.0
	}
}

func FilterByLabel(families []*MetricFamily, labelName, labelValue string) []*MetricFamily {
	var result []*MetricFamily
	for _, family := range families {
		var filteredMetrics []*Metric
		for _, metric := range family.Metrics {
			if val, ok := metric.Labels[labelName]; ok && val == labelValue {
				filteredMetrics = append(filteredMetrics, metric)
			}
		}
		if len(filteredMetrics) > 0 {
			result = append(result, &MetricFamily{
				Name:    family.Name,
				Help:    family.Help,
				Type:    family.Type,
				Metrics: filteredMetrics,
			})
		}
	}
	return result
}

func FindFamilyByName(families []*MetricFamily, name string) *MetricFamily {
	for _, family := range families {
		if family.Name == name {
			return family
		}
	}
	return nil
}

func (mf *MetricFamily) Sum() float64 {
	var sum float64
	for _, m := range mf.Metrics {
		sum += m.Value
	}
	return sum
}

func (mf *MetricFamily) Avg() float64 {
	if len(mf.Metrics) == 0 {
		return 0.0
	}
	return mf.Sum() / float64(len(mf.Metrics))
}

func (mf *MetricFamily) Min() float64 {
	if len(mf.Metrics) == 0 {
		return 0.0
	}
	min := mf.Metrics[0].Value
	for _, m := range mf.Metrics[1:] {
		if m.Value < min {
			min = m.Value
		}
	}
	return min
}

func (mf *MetricFamily) Max() float64 {
	if len(mf.Metrics) == 0 {
		return 0.0
	}
	max := mf.Metrics[0].Value
	for _, m := range mf.Metrics[1:] {
		if m.Value > max {
			max = m.Value
		}
	}
	return max
}

func (mf *MetricFamily) String() string {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	fmt.Fprintf(writer, "# HELP %s %s\n", mf.Name, mf.Help)
	fmt.Fprintf(writer, "# TYPE %s %s\n", mf.Name, mf.Type.String())

	for _, m := range mf.Metrics {
		fmt.Fprintf(writer, "%s", mf.Name)
		if len(m.Labels) > 0 {
			fmt.Fprint(writer, "{")
			first := true
			for k, v := range m.Labels {
				if !first {
					fmt.Fprint(writer, ",")
				}
				fmt.Fprintf(writer, "%s=\"%s\"", k, v)
				first = false
			}
			fmt.Fprint(writer, "}")
		}
		fmt.Fprintf(writer, " %v", m.Value)
		if m.Timestamp > 0 {
			fmt.Fprintf(writer, " %d", m.Timestamp)
		}
		fmt.Fprintln(writer)
	}

	writer.Flush()
	return buf.String()
}

func RegisterCollectors(collectors ...prometheus.Collector) error {
	for _, c := range collectors {
		if err := prometheus.Register(c); err != nil {
			return fmt.Errorf("failed to register collector: %w", err)
		}
	}
	return nil
}
