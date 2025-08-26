package collector

import (
	"context"
)

type Collector interface {
	Name() string
	Priority() int
	Collect(ctx context.Context) (map[string]interface{}, error)
}

type BaseCollector struct {
	name     string
	priority int
}

func NewBaseCollector(name string, priority int) BaseCollector {
	return BaseCollector{
		name:     name,
		priority: priority,
	}
}

func (b BaseCollector) Name() string {
	return b.name
}

func (b BaseCollector) Priority() int {
	return b.priority
}

func (b BaseCollector) Collect(_ context.Context) (map[string]interface{}, error) {
	return nil, nil
}
