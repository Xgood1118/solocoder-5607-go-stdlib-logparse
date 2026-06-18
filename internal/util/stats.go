package util

import (
	"sort"
)

type FloatSlice []float64

func (s FloatSlice) Len() int           { return len(s) }
func (s FloatSlice) Less(i, j int) bool { return s[i] < s[j] }
func (s FloatSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := p / 100.0 * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	frac := index - float64(lower)
	return sorted[lower] + frac*(sorted[upper]-sorted[lower])
}

type IntCounter struct {
	Total   int64
	Count   int64
	Min     float64
	Max     float64
	Sum     float64
	Values  []float64
	MaxKeep int
}

func NewIntCounter(maxKeep int) *IntCounter {
	return &IntCounter{
		MaxKeep: maxKeep,
		Min:     1e18,
	}
}

func (c *IntCounter) Add(v float64) {
	c.Total++
	c.Count++
	c.Sum += v
	if v < c.Min {
		c.Min = v
	}
	if v > c.Max {
		c.Max = v
	}
	if c.MaxKeep > 0 && len(c.Values) < c.MaxKeep {
		c.Values = append(c.Values, v)
	}
}

func (c *IntCounter) Avg() float64 {
	if c.Count == 0 {
		return 0
	}
	return c.Sum / float64(c.Count)
}

func (c *IntCounter) Percentile(p float64) float64 {
	if len(c.Values) == 0 {
		return c.Avg()
	}
	return Percentile(c.Values, p)
}
