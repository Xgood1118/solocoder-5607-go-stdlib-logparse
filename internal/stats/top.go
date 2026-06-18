package stats

import (
	"sort"
)

type StringCount struct {
	Key   string
	Count int64
}

type PathStatItem struct {
	Path       string
	Count      int64
	ErrorCount int64
	AvgTime    float64
	MaxTime    float64
	MinTime    float64
	ErrorRate  float64
}

func (a *Aggregator) TopIPs(top int) []StringCount {
	return topN(a.IPs, top)
}

func (a *Aggregator) TopPaths(top int, sortBy string) []PathStatItem {
	items := make([]PathStatItem, 0, len(a.Paths))
	for path, ps := range a.Paths {
		avg := 0.0
		if ps.Count > 0 && ps.TotalTime > 0 {
			avg = ps.TotalTime / float64(ps.Count)
		}
		errRate := 0.0
		if ps.Count > 0 {
			errRate = float64(ps.ErrorCount) / float64(ps.Count) * 100
		}
		items = append(items, PathStatItem{
			Path:       path,
			Count:      ps.Count,
			ErrorCount: ps.ErrorCount,
			AvgTime:    avg,
			MaxTime:    ps.MaxTime,
			MinTime:    ps.MinTime,
			ErrorRate:  errRate,
		})
	}

	switch sortBy {
	case "time":
		sort.Slice(items, func(i, j int) bool {
			return items[i].AvgTime > items[j].AvgTime
		})
	case "error":
		sort.Slice(items, func(i, j int) bool {
			return items[i].ErrorRate > items[j].ErrorRate
		})
	default:
		sort.Slice(items, func(i, j int) bool {
			return items[i].Count > items[j].Count
		})
	}

	if top > 0 && top < len(items) {
		items = items[:top]
	}
	return items
}

func (a *Aggregator) TopReferers(top int) []StringCount {
	return topN(a.Referers, top)
}

func (a *Aggregator) TopUserAgents(top int) []StringCount {
	return topN(a.UserAgents, top)
}

func (a *Aggregator) TopHours(top int) []StringCount {
	return topN(a.Hours, top)
}

func (a *Aggregator) TopDays(top int) []StringCount {
	return topN(a.Days, top)
}

func (a *Aggregator) StatusDistribution() []StringCount {
	result := make([]StringCount, 0, len(a.StatusCodes))
	keys := make([]int, 0, len(a.StatusCodes))
	for k := range a.StatusCodes {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		result = append(result, StringCount{
			Key:   itoa(k),
			Count: a.StatusCodes[k],
		})
	}
	return result
}

func topN(m map[string]int64, top int) []StringCount {
	items := make([]StringCount, 0, len(m))
	for k, v := range m {
		items = append(items, StringCount{Key: k, Count: v})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})
	if top > 0 && top < len(items) {
		items = items[:top]
	}
	return items
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
