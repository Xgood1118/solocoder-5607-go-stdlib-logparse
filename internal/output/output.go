package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"logparse/internal/stats"
	"strconv"
	"text/tabwriter"
)

type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatCSV   OutputFormat = "csv"
)

type OutputOptions struct {
	Format   OutputFormat
	Verbose  bool
	TopN     int
	SortBy   string
}

type OutputResult struct {
	Summary SummaryResult    `json:"summary"`
	Status  []StatusItem     `json:"status_codes"`
	Paths   []PathResultItem `json:"top_paths"`
	IPs     []IPResultItem   `json:"top_ips"`
	Hours   []TimeResultItem `json:"top_hours"`
	Referer []RefererResult  `json:"top_referers"`
	UA      []UAResult       `json:"top_user_agents"`
}

type SummaryResult struct {
	TotalRequests  int64   `json:"total_requests"`
	ErrorRequests int64   `json:"error_requests"`
	ErrorRate     float64 `json:"error_rate"`
	TotalBytes    int64   `json:"total_bytes"`
	AvgRespTime   float64 `json:"avg_resp_time"`
	P50RespTime   float64 `json:"p50_resp_time"`
	P95RespTime   float64 `json:"p95_resp_time"`
	P99RespTime   float64 `json:"p99_resp_time"`
	MaxRespTime   float64 `json:"max_resp_time"`
	MinRespTime   float64 `json:"min_resp_time"`
	FailedLines   int64   `json:"failed_lines"`
	TotalLines    int64   `json:"total_lines"`
}

type StatusItem struct {
	Code  string `json:"code"`
	Count int64  `json:"count"`
}

type PathResultItem struct {
	Path       string  `json:"path"`
	Count      int64   `json:"count"`
	ErrorCount int64   `json:"error_count"`
	ErrorRate  float64 `json:"error_rate"`
	AvgTime    float64 `json:"avg_time"`
}

type IPResultItem struct {
	IP    string `json:"ip"`
	Count int64  `json:"count"`
}

type TimeResultItem struct {
	Time  string `json:"time"`
	Count int64  `json:"count"`
}

type RefererResult struct {
	Referer string `json:"referer"`
	Count   int64  `json:"count"`
}

type UAResult struct {
	UserAgent string `json:"user_agent"`
	Count     int64  `json:"count"`
}

func Write(w io.Writer, agg *stats.Aggregator, opts OutputOptions) error {
	switch opts.Format {
	case FormatJSON:
		return writeJSON(w, agg, opts)
	case FormatCSV:
		return writeCSV(w, agg, opts)
	default:
		return writeTable(w, agg, opts)
	}
}

func buildResult(agg *stats.Aggregator, opts OutputOptions) OutputResult {
	topN := opts.TopN
	if topN <= 0 {
		topN = 10
	}

	paths := agg.TopPaths(topN, opts.SortBy)
	pathItems := make([]PathResultItem, len(paths))
	for i, p := range paths {
		pathItems[i] = PathResultItem{
			Path:       p.Path,
			Count:      p.Count,
			ErrorCount: p.ErrorCount,
			ErrorRate:  p.ErrorRate,
			AvgTime:    p.AvgTime,
		}
	}

	ips := agg.TopIPs(topN)
	ipItems := make([]IPResultItem, len(ips))
	for i, ip := range ips {
		ipItems[i] = IPResultItem{IP: ip.Key, Count: ip.Count}
	}

	hours := agg.TopHours(topN)
	hourItems := make([]TimeResultItem, len(hours))
	for i, h := range hours {
		hourItems[i] = TimeResultItem{Time: h.Key, Count: h.Count}
	}

	refs := agg.TopReferers(topN)
	refItems := make([]RefererResult, len(refs))
	for i, r := range refs {
		refItems[i] = RefererResult{Referer: r.Key, Count: r.Count}
	}

	uas := agg.TopUserAgents(topN)
	uaItems := make([]UAResult, len(uas))
	for i, ua := range uas {
		uaItems[i] = UAResult{UserAgent: ua.Key, Count: ua.Count}
	}

	statusDist := agg.StatusDistribution()
	statusItems := make([]StatusItem, len(statusDist))
	for i, s := range statusDist {
		statusItems[i] = StatusItem{Code: s.Key, Count: s.Count}
	}

	return OutputResult{
		Summary: SummaryResult{
			TotalRequests: agg.TotalRequests,
			ErrorRequests: agg.ErrorRequests,
			ErrorRate:     agg.ErrorRate(),
			TotalBytes:    agg.TotalBytes,
			AvgRespTime:   agg.AvgRespTime(),
			P50RespTime:  agg.P50(),
			P95RespTime:  agg.P95(),
			P99RespTime:  agg.P99(),
			MaxRespTime:  agg.MaxResp,
			MinRespTime:  agg.MinRespTime(),
			FailedLines:  agg.FailedLines,
			TotalLines:   agg.TotalLines,
		},
		Status:  statusItems,
		Paths:   pathItems,
		IPs:     ipItems,
		Hours:   hourItems,
		Referer: refItems,
		UA:      uaItems,
	}
}

func writeJSON(w io.Writer, agg *stats.Aggregator, opts OutputOptions) error {
	result := buildResult(agg, opts)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func writeCSV(w io.Writer, agg *stats.Aggregator, opts OutputOptions) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	if opts.Verbose {
		_ = cw.Write([]string{"Section", "Metric", "Value"})
		_ = cw.Write([]string{"Summary", "Total Requests", strconv.FormatInt(agg.TotalRequests, 10)})
		_ = cw.Write([]string{"Summary", "Error Requests", strconv.FormatInt(agg.ErrorRequests, 10)})
		_ = cw.Write([]string{"Summary", "Error Rate (%)", fmt.Sprintf("%.2f", agg.ErrorRate())})
		_ = cw.Write([]string{"Summary", "Avg Resp Time", fmt.Sprintf("%.3f", agg.AvgRespTime())})
		_ = cw.Write([]string{"Summary", "P50 Resp Time", fmt.Sprintf("%.3f", agg.P50())})
		_ = cw.Write([]string{"Summary", "P95 Resp Time", fmt.Sprintf("%.3f", agg.P95())})
		_ = cw.Write([]string{"Summary", "P99 Resp Time", fmt.Sprintf("%.3f", agg.P99())})
		_ = cw.Write([]string{"Summary", "Max Resp Time", fmt.Sprintf("%.3f", agg.MaxResp)})
		_ = cw.Write([]string{"Summary", "Failed Lines", strconv.FormatInt(agg.FailedLines, 10)})

		_ = cw.Write([]string{"", "", ""})
		_ = cw.Write([]string{"Top Paths", "Path", "Count", "Avg Time", "Error Rate"})
		topN := opts.TopN
		if topN <= 0 {
			topN = 10
		}
		for _, p := range agg.TopPaths(topN, opts.SortBy) {
			_ = cw.Write([]string{"", p.Path, strconv.FormatInt(p.Count, 10),
				fmt.Sprintf("%.3f", p.AvgTime), fmt.Sprintf("%.2f%%", p.ErrorRate)})
		}
	} else {
		_ = cw.Write([]string{"path", "count", "avg_time"})
		topN := opts.TopN
		if topN <= 0 {
			topN = 10
		}
		for _, p := range agg.TopPaths(topN, opts.SortBy) {
			_ = cw.Write([]string{p.Path, strconv.FormatInt(p.Count, 10), fmt.Sprintf("%.3f", p.AvgTime)})
		}
	}
	return nil
}

func writeTable(w io.Writer, agg *stats.Aggregator, opts OutputOptions) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	if opts.Verbose {
		fmt.Fprintln(tw, "=== Summary ===")
		fmt.Fprintf(tw, "Total Requests:\t%d\n", agg.TotalRequests)
		fmt.Fprintf(tw, "Error Requests:\t%d\n", agg.ErrorRequests)
		fmt.Fprintf(tw, "Error Rate:\t%.2f%%\n", agg.ErrorRate())
		fmt.Fprintf(tw, "Total Bytes:\t%d\n", agg.TotalBytes)
		fmt.Fprintf(tw, "Avg Resp Time:\t%.3f\n", agg.AvgRespTime())
		fmt.Fprintf(tw, "P50 Resp Time:\t%.3f\n", agg.P50())
		fmt.Fprintf(tw, "P95 Resp Time:\t%.3f\n", agg.P95())
		fmt.Fprintf(tw, "P99 Resp Time:\t%.3f\n", agg.P99())
		fmt.Fprintf(tw, "Max Resp Time:\t%.3f\n", agg.MaxResp)
		fmt.Fprintf(tw, "Min Resp Time:\t%.3f\n", agg.MinRespTime())
		if agg.FailedLines > 0 {
			fmt.Fprintf(tw, "Failed Lines:\t%d / %d\n", agg.FailedLines, agg.TotalLines)
		}
		fmt.Fprintln(tw)

		fmt.Fprintln(tw, "=== Status Code Distribution ===")
		for _, s := range agg.StatusDistribution() {
			fmt.Fprintf(tw, "%s:\t%d\n", s.Key, s.Count)
		}
		fmt.Fprintln(tw)

		topN := opts.TopN
		if topN <= 0 {
			topN = 10
		}

		fmt.Fprintf(tw, "=== Top %d Paths ===\n", topN)
		fmt.Fprintln(tw, "Path\tCount\tAvg Time\tError Rate")
		for _, p := range agg.TopPaths(topN, opts.SortBy) {
			fmt.Fprintf(tw, "%s\t%d\t%.3f\t%.2f%%\n", p.Path, p.Count, p.AvgTime, p.ErrorRate)
		}
		fmt.Fprintln(tw)

		fmt.Fprintf(tw, "=== Top %d IPs ===\n", topN)
		fmt.Fprintln(tw, "IP\tCount")
		for _, ip := range agg.TopIPs(topN) {
			fmt.Fprintf(tw, "%s\t%d\n", ip.Key, ip.Count)
		}
		fmt.Fprintln(tw)

		fmt.Fprintf(tw, "=== Top %d Hours ===\n", topN)
		fmt.Fprintln(tw, "Hour\tCount")
		for _, h := range agg.TopHours(topN) {
			fmt.Fprintf(tw, "%s\t%d\n", h.Key, h.Count)
		}
		fmt.Fprintln(tw)

		if len(agg.Referers) > 0 {
			fmt.Fprintf(tw, "=== Top %d Referers ===\n", topN)
			fmt.Fprintln(tw, "Referer\tCount")
			for _, r := range agg.TopReferers(topN) {
				fmt.Fprintf(tw, "%s\t%d\n", r.Key, r.Count)
			}
			fmt.Fprintln(tw)
		}

		if len(agg.UserAgents) > 0 {
			fmt.Fprintf(tw, "=== Top %d User-Agents ===\n", topN)
			fmt.Fprintln(tw, "User-Agent\tCount")
			for _, ua := range agg.TopUserAgents(topN) {
				fmt.Fprintf(tw, "%s\t%d\n", ua.Key, ua.Count)
			}
			fmt.Fprintln(tw)
		}
	} else {
		topN := opts.TopN
		if topN <= 0 {
			topN = 10
		}
		for _, p := range agg.TopPaths(topN, opts.SortBy) {
			fmt.Fprintf(tw, "%s\t%d\t%.3f\n", p.Path, p.Count, p.AvgTime)
		}
	}

	return nil
}
