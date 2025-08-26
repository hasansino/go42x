package collector

import (
	"context"
	"fmt"
	"sort"
)

// ContextManager orchestrates context collection from multiple collectors
type ContextManager struct {
	collectors []Collector
}

// NewManager creates a new context collection manager
func NewManager() *ContextManager {
	return &ContextManager{
		collectors: make([]Collector, 0),
	}
}

// RegisterCollector adds a new collector to the manager
func (m *ContextManager) RegisterCollector(c Collector) {
	m.collectors = append(m.collectors, c)
}

// RegisterCollectors adds multiple collectors at once
func (m *ContextManager) RegisterCollectors(collectors ...Collector) {
	for _, c := range collectors {
		m.RegisterCollector(c)
	}
}

// Collect runs all registered collectors and returns the combined context
func (m *ContextManager) Collect(ctx context.Context) (map[string]interface{}, error) {
	collectors := make([]Collector, len(m.collectors))
	copy(collectors, m.collectors)

	sort.Slice(collectors, func(i, j int) bool {
		return collectors[i].Priority() < collectors[j].Priority()
	})

	result := make(map[string]interface{})

	for _, c := range collectors {
		data, err := c.Collect(ctx)
		if err != nil {
			return nil, fmt.Errorf("collector %s failed: %w", c.Name(), err)
		}
		if len(data) > 0 {
			result[c.Name()] = data
		}
	}

	return result, nil
}
