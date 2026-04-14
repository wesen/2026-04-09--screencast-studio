package metrics

import (
	"strings"
	"testing"
)

func TestRegistryWritePrometheusRendersCountersAndGauges(t *testing.T) {
	registry := NewRegistry()
	counter := registry.MustRegisterCounterVec(
		"test_counter_total",
		"Counter help.",
		"kind",
	)
	gauge := registry.MustRegisterGaugeVec(
		"test_gauge",
		"Gauge help.",
		"kind",
	)

	counter.Add(map[string]string{"kind": "alpha"}, 3)
	gauge.Set(map[string]string{"kind": "beta"}, 7)
	gauge.Dec(map[string]string{"kind": "beta"})

	var b strings.Builder
	if err := registry.WritePrometheus(&b); err != nil {
		t.Fatalf("WritePrometheus() error = %v", err)
	}
	body := b.String()
	if !strings.Contains(body, "# TYPE test_counter_total counter") {
		t.Fatalf("missing counter type line: %s", body)
	}
	if !strings.Contains(body, "test_counter_total{kind=\"alpha\"} 3") {
		t.Fatalf("missing counter sample: %s", body)
	}
	if !strings.Contains(body, "# TYPE test_gauge gauge") {
		t.Fatalf("missing gauge type line: %s", body)
	}
	if !strings.Contains(body, "test_gauge{kind=\"beta\"} 6") {
		t.Fatalf("missing gauge sample: %s", body)
	}
}
