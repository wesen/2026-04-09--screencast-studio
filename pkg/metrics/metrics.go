package metrics

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type Registry struct {
	mu       sync.RWMutex
	counters map[string]*CounterVec
	gauges   map[string]*GaugeVec
}

type CounterVec struct {
	name       string
	help       string
	labelNames []string

	mu     sync.RWMutex
	values map[string]*counterEntry
}

type GaugeVec struct {
	name       string
	help       string
	labelNames []string

	mu     sync.RWMutex
	values map[string]*gaugeEntry
}

type counterEntry struct {
	labels map[string]string
	value  atomic.Uint64
}

type gaugeEntry struct {
	labels map[string]string
	value  atomic.Int64
}

type metricFamily struct {
	name     string
	help     string
	typeName string
	render   func(io.Writer) error
}

var defaultRegistry = NewRegistry()

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func NewRegistry() *Registry {
	return &Registry{
		counters: map[string]*CounterVec{},
		gauges:   map[string]*GaugeVec{},
	}
}

func MustRegisterCounterVec(name, help string, labelNames ...string) *CounterVec {
	return defaultRegistry.MustRegisterCounterVec(name, help, labelNames...)
}

func MustRegisterGaugeVec(name, help string, labelNames ...string) *GaugeVec {
	return defaultRegistry.MustRegisterGaugeVec(name, help, labelNames...)
}

func (r *Registry) MustRegisterCounterVec(name, help string, labelNames ...string) *CounterVec {
	if r == nil {
		panic("metrics registry is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing := r.counters[name]; existing != nil {
		if !sameStrings(existing.labelNames, labelNames) {
			panic(fmt.Sprintf("metrics counter %s already registered with different label names", name))
		}
		return existing
	}
	if existing := r.gauges[name]; existing != nil {
		panic(fmt.Sprintf("metrics name %s already registered as gauge", name))
	}
	counter := &CounterVec{
		name:       name,
		help:       help,
		labelNames: append([]string(nil), labelNames...),
		values:     map[string]*counterEntry{},
	}
	r.counters[name] = counter
	return counter
}

func (r *Registry) MustRegisterGaugeVec(name, help string, labelNames ...string) *GaugeVec {
	if r == nil {
		panic("metrics registry is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing := r.gauges[name]; existing != nil {
		if !sameStrings(existing.labelNames, labelNames) {
			panic(fmt.Sprintf("metrics gauge %s already registered with different label names", name))
		}
		return existing
	}
	if existing := r.counters[name]; existing != nil {
		panic(fmt.Sprintf("metrics name %s already registered as counter", name))
	}
	gauge := &GaugeVec{
		name:       name,
		help:       help,
		labelNames: append([]string(nil), labelNames...),
		values:     map[string]*gaugeEntry{},
	}
	r.gauges[name] = gauge
	return gauge
}

func (r *Registry) WritePrometheus(w io.Writer) error {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	families := make([]metricFamily, 0, len(r.counters)+len(r.gauges))
	for _, counter := range r.counters {
		metric := counter
		families = append(families, metricFamily{
			name:     metric.name,
			help:     metric.help,
			typeName: "counter",
			render: func(w io.Writer) error {
				entries := metric.snapshot()
				for _, entry := range entries {
					if _, err := fmt.Fprintf(w, "%s%s %d\n", metric.name, formatLabels(metric.labelNames, entry.labels), entry.value.Load()); err != nil {
						return err
					}
				}
				return nil
			},
		})
	}
	for _, gauge := range r.gauges {
		metric := gauge
		families = append(families, metricFamily{
			name:     metric.name,
			help:     metric.help,
			typeName: "gauge",
			render: func(w io.Writer) error {
				entries := metric.snapshot()
				for _, entry := range entries {
					if _, err := fmt.Fprintf(w, "%s%s %d\n", metric.name, formatLabels(metric.labelNames, entry.labels), entry.value.Load()); err != nil {
						return err
					}
				}
				return nil
			},
		})
	}
	r.mu.RUnlock()

	sort.Slice(families, func(i, j int) bool { return families[i].name < families[j].name })
	for _, family := range families {
		if _, err := fmt.Fprintf(w, "# HELP %s %s\n", family.name, escapeHelp(family.help)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "# TYPE %s %s\n", family.name, family.typeName); err != nil {
			return err
		}
		if err := family.render(w); err != nil {
			return err
		}
	}
	return nil
}

func (c *CounterVec) Inc(labels map[string]string) {
	c.Add(labels, 1)
}

func (c *CounterVec) Add(labels map[string]string, delta uint64) {
	if c == nil {
		return
	}
	entry := c.entry(labels)
	entry.value.Add(delta)
}

func (c *CounterVec) entry(labels map[string]string) *counterEntry {
	key, normalized := normalizeLabels(c.labelNames, labels)
	c.mu.RLock()
	existing := c.values[key]
	c.mu.RUnlock()
	if existing != nil {
		return existing
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if existing = c.values[key]; existing != nil {
		return existing
	}
	entry := &counterEntry{labels: normalized}
	c.values[key] = entry
	return entry
}

func (c *CounterVec) snapshot() []*counterEntry {
	c.mu.RLock()
	entries := make([]*counterEntry, 0, len(c.values))
	for _, entry := range c.values {
		entries = append(entries, entry)
	}
	c.mu.RUnlock()
	sort.Slice(entries, func(i, j int) bool {
		return labelSetKey(c.labelNames, entries[i].labels) < labelSetKey(c.labelNames, entries[j].labels)
	})
	return entries
}

func (g *GaugeVec) Inc(labels map[string]string) {
	g.Add(labels, 1)
}

func (g *GaugeVec) Dec(labels map[string]string) {
	g.Add(labels, -1)
}

func (g *GaugeVec) Set(labels map[string]string, value int64) {
	if g == nil {
		return
	}
	entry := g.entry(labels)
	entry.value.Store(value)
}

func (g *GaugeVec) Add(labels map[string]string, delta int64) {
	if g == nil {
		return
	}
	entry := g.entry(labels)
	entry.value.Add(delta)
}

func (g *GaugeVec) entry(labels map[string]string) *gaugeEntry {
	key, normalized := normalizeLabels(g.labelNames, labels)
	g.mu.RLock()
	existing := g.values[key]
	g.mu.RUnlock()
	if existing != nil {
		return existing
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if existing = g.values[key]; existing != nil {
		return existing
	}
	entry := &gaugeEntry{labels: normalized}
	g.values[key] = entry
	return entry
}

func (g *GaugeVec) snapshot() []*gaugeEntry {
	g.mu.RLock()
	entries := make([]*gaugeEntry, 0, len(g.values))
	for _, entry := range g.values {
		entries = append(entries, entry)
	}
	g.mu.RUnlock()
	sort.Slice(entries, func(i, j int) bool {
		return labelSetKey(g.labelNames, entries[i].labels) < labelSetKey(g.labelNames, entries[j].labels)
	})
	return entries
}

func normalizeLabels(labelNames []string, labels map[string]string) (string, map[string]string) {
	normalized := make(map[string]string, len(labelNames))
	parts := make([]string, 0, len(labelNames))
	for _, name := range labelNames {
		value := ""
		if labels != nil {
			value = labels[name]
		}
		normalized[name] = value
		parts = append(parts, name+"="+value)
	}
	return strings.Join(parts, "\xff"), normalized
}

func formatLabels(labelNames []string, labels map[string]string) string {
	if len(labelNames) == 0 {
		return ""
	}
	parts := make([]string, 0, len(labelNames))
	for _, name := range labelNames {
		parts = append(parts, fmt.Sprintf("%s=%s", name, strconv.Quote(labels[name])))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func labelSetKey(labelNames []string, labels map[string]string) string {
	parts := make([]string, 0, len(labelNames))
	for _, name := range labelNames {
		parts = append(parts, name+"="+labels[name])
	}
	return strings.Join(parts, "\xff")
}

func escapeHelp(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
