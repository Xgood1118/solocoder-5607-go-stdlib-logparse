package stats

import (
	"logparse/internal/util"
	"sync"
	"time"
)

type Aggregator struct {
	mu sync.Mutex

	TotalRequests int64
	ErrorRequests int64
	TotalBytes    int64

	RespTimes  []float64
	MaxResp    float64
	MinResp    float64
	SumResp    float64
	maxSamples int

	StatusCodes map[int]int64
	Paths       map[string]*PathStat
	IPs         map[string]int64
	Hours       map[string]int64
	Days        map[string]int64
	Referers    map[string]int64
	UserAgents  map[string]int64

	StartTime time.Time
	EndTime   time.Time

	FailedLines int64
	TotalLines  int64
}

type PathStat struct {
	Count      int64
	ErrorCount int64
	TotalTime  float64
	MaxTime    float64
	MinTime    float64
}

func NewAggregator(maxSamples int) *Aggregator {
	if maxSamples <= 0 {
		maxSamples = 10000
	}
	return &Aggregator{
		StatusCodes: make(map[int]int64),
		Paths:       make(map[string]*PathStat),
		IPs:         make(map[string]int64),
		Hours:       make(map[string]int64),
		Days:        make(map[string]int64),
		Referers:    make(map[string]int64),
		UserAgents:  make(map[string]int64),
		MinResp:     1e18,
		maxSamples:  maxSamples,
	}
}

func (a *Aggregator) Add(entry *util.LogEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.TotalRequests++

	if util.IsErrorStatus(entry.Status) {
		a.ErrorRequests++
	}

	a.TotalBytes += entry.Bytes

	if entry.Status > 0 {
		a.StatusCodes[entry.Status]++
	}

	if entry.Path != "" {
		ps, ok := a.Paths[entry.Path]
		if !ok {
			ps = &PathStat{MinTime: 1e18}
			a.Paths[entry.Path] = ps
		}
		ps.Count++
		if util.IsErrorStatus(entry.Status) {
			ps.ErrorCount++
		}
		if entry.RespTime > 0 {
			ps.TotalTime += entry.RespTime
			if entry.RespTime > ps.MaxTime {
				ps.MaxTime = entry.RespTime
			}
			if entry.RespTime < ps.MinTime {
				ps.MinTime = entry.RespTime
			}
		}
	}

	if entry.IP != "" {
		a.IPs[entry.IP]++
	}

	if !entry.Time.IsZero() {
		hourKey := util.TruncateToHour(entry.Time).Format("2006-01-02 15:00")
		a.Hours[hourKey]++
		dayKey := util.TruncateToDay(entry.Time).Format("2006-01-02")
		a.Days[dayKey]++

		if a.StartTime.IsZero() || entry.Time.Before(a.StartTime) {
			a.StartTime = entry.Time
		}
		if a.EndTime.IsZero() || entry.Time.After(a.EndTime) {
			a.EndTime = entry.Time
		}
	}

	if entry.Referer != "" && entry.Referer != "-" {
		a.Referers[entry.Referer]++
	}

	if entry.UserAgent != "" && entry.UserAgent != "-" {
		a.UserAgents[entry.UserAgent]++
	}

	if entry.RespTime > 0 {
		a.SumResp += entry.RespTime
		if entry.RespTime > a.MaxResp {
			a.MaxResp = entry.RespTime
		}
		if entry.RespTime < a.MinResp {
			a.MinResp = entry.RespTime
		}
		if len(a.RespTimes) < a.maxSamples {
			a.RespTimes = append(a.RespTimes, entry.RespTime)
		}
	}
}

func (a *Aggregator) AddFailed() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.FailedLines++
	a.TotalLines++
}

func (a *Aggregator) AddTotalLine() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.TotalLines++
}

func (a *Aggregator) ErrorRate() float64 {
	if a.TotalRequests == 0 {
		return 0
	}
	return float64(a.ErrorRequests) / float64(a.TotalRequests) * 100
}

func (a *Aggregator) AvgRespTime() float64 {
	if a.TotalRequests == 0 || a.SumResp == 0 {
		return 0
	}
	return a.SumResp / float64(a.TotalRequests)
}

func (a *Aggregator) P95() float64 {
	return util.Percentile(a.RespTimes, 95)
}

func (a *Aggregator) P99() float64 {
	return util.Percentile(a.RespTimes, 99)
}

func (a *Aggregator) MinRespTime() float64 {
	if a.MinResp == 1e18 {
		return 0
	}
	return a.MinResp
}
