package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"logparse/internal/filter"
	"logparse/internal/output"
	"logparse/internal/parser"
	"logparse/internal/stats"
	"logparse/internal/util"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var (
		files      stringSlice
		dirPath    string
		recursive  bool
		format     string
		pattern    string
		fromStr    string
		toStr      string
		statusList stringSlice
		pathPat    string
		ipPat      string
		method     string
		uaStr      string
		outFormat  string
		jsonOut    bool
		csvOut     bool
		verbose    bool
		topN       int
		sortBy     string
		progress   bool
	)

	flag.Var(&files, "f", "日志文件路径，可重复指定多个")
	flag.StringVar(&dirPath, "d", "", "日志目录路径")
	flag.BoolVar(&recursive, "r", false, "递归读取目录下的日志文件")
	flag.StringVar(&format, "format", "auto", "日志格式：auto, combined, error, json, regex")
	flag.StringVar(&pattern, "pattern", "", "自定义正则表达式（用于 regex 格式）")
	flag.StringVar(&fromStr, "from", "", "开始时间（YYYY-MM-DD 或 YYYY-MM-DD HH:MM:SS）")
	flag.StringVar(&toStr, "to", "", "结束时间（YYYY-MM-DD 或 YYYY-MM-DD HH:MM:SS）")
	flag.Var(&statusList, "status", "状态码过滤，如 200 4xx 5xx，可重复")
	flag.StringVar(&pathPat, "path", "", "请求路径过滤（支持通配符）")
	flag.StringVar(&ipPat, "ip", "", "IP 过滤（支持通配符和 CIDR）")
	flag.StringVar(&method, "method", "", "HTTP 方法过滤")
	flag.StringVar(&uaStr, "ua", "", "User-Agent 包含字符串过滤")
	flag.BoolVar(&jsonOut, "json", false, "JSON 格式输出")
	flag.BoolVar(&csvOut, "csv", false, "CSV 格式输出")
	flag.StringVar(&outFormat, "out", "table", "输出格式：table, json, csv")
	flag.BoolVar(&verbose, "verbose", false, "详细输出模式")
	flag.IntVar(&topN, "top", 10, "Top N 数量")
	flag.StringVar(&sortBy, "sort", "count", "排序方式：count, time, error")
	flag.BoolVar(&progress, "progress", false, "显示进度条")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "logparse - 日志分析工具\n\n")
		fmt.Fprintf(os.Stderr, "用法：logparse [选项]\n\n")
		fmt.Fprintf(os.Stderr, "输入方式：\n")
		fmt.Fprintf(os.Stderr, "  -f file       日志文件，可重复指定\n")
		fmt.Fprintf(os.Stderr, "  -d dir        日志目录\n")
		fmt.Fprintf(os.Stderr, "  -r            递归读取目录\n")
		fmt.Fprintf(os.Stderr, "  stdin         管道输入：cat *.log | logparse\n\n")
		fmt.Fprintf(os.Stderr, "日志格式：\n")
		fmt.Fprintf(os.Stderr, "  -format       auto|combined|error|json|regex\n")
		fmt.Fprintf(os.Stderr, "  -pattern      自定义正则（regex 格式用）\n\n")
		fmt.Fprintf(os.Stderr, "过滤条件：\n")
		fmt.Fprintf(os.Stderr, "  -from         开始时间\n")
		fmt.Fprintf(os.Stderr, "  -to           结束时间\n")
		fmt.Fprintf(os.Stderr, "  -status       状态码，如 4xx 5xx\n")
		fmt.Fprintf(os.Stderr, "  -path         请求路径\n")
		fmt.Fprintf(os.Stderr, "  -ip           IP 地址\n")
		fmt.Fprintf(os.Stderr, "  -method       HTTP 方法\n")
		fmt.Fprintf(os.Stderr, "  -ua           User-Agent 包含\n\n")
		fmt.Fprintf(os.Stderr, "输出选项：\n")
		fmt.Fprintf(os.Stderr, "  -json         JSON 输出\n")
		fmt.Fprintf(os.Stderr, "  -csv          CSV 输出\n")
		fmt.Fprintf(os.Stderr, "  -verbose      详细输出\n")
		fmt.Fprintf(os.Stderr, "  -top N        Top N\n")
		fmt.Fprintf(os.Stderr, "  -sort         count|time|error\n")
		fmt.Fprintf(os.Stderr, "  -progress     显示进度条\n")
	}

	flag.Parse()

	fil := &filter.Filter{
		StatusCodes: statusList,
		PathPattern: pathPat,
		IPPattern:   ipPat,
		Method:      method,
		UserAgent:   uaStr,
	}

	if fromStr != "" {
		t, err := util.ParseDateWithTimezone(fromStr, time.Local)
		if err != nil {
			fmt.Fprintf(os.Stderr, "无效的 -from 时间: %v\n", err)
			os.Exit(1)
		}
		fil.From = t
	}
	if toStr != "" {
		t, err := util.ParseDateWithTimezone(toStr, time.Local)
		if err != nil {
			fmt.Fprintf(os.Stderr, "无效的 -to 时间: %v\n", err)
			os.Exit(1)
		}
		fil.To = t.AddDate(0, 0, 1).Add(-1)
	}

	createParser := func() (parser.Parser, error) {
		switch strings.ToLower(format) {
		case "combined", "access", "nginx", "apache":
			return parser.NewCombinedParser(), nil
		case "error", "nginx-error", "apache-error":
			return parser.NewErrorParser(), nil
		case "json", "jsonl":
			return parser.NewJSONParser(), nil
		case "regex", "custom":
			if pattern == "" {
				return nil, fmt.Errorf("regex 格式需要指定 -pattern 参数")
			}
			return parser.NewRegexParser(pattern)
		default:
			return parser.NewAutoParser(), nil
		}
	}

	p, err := createParser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	agg := stats.NewAggregator(100000)

	logFiles := collectFiles(files, dirPath, recursive)

	useStdin := len(logFiles) == 0
	if useStdin {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			flag.Usage()
			os.Exit(1)
		}
	}

	outFmt := output.FormatTable
	if jsonOut || outFormat == "json" {
		outFmt = output.FormatJSON
	} else if csvOut || outFormat == "csv" {
		outFmt = output.FormatCSV
	}

	if useStdin {
		processReader(os.Stdin, p, fil, agg, "stdin", progress, 0)
	} else {
		processFilesParallel(logFiles, createParser, fil, agg, progress)
	}

	opts := output.OutputOptions{
		Format:  outFmt,
		Verbose: verbose,
		TopN:    topN,
		SortBy:  sortBy,
	}

	if err := output.Write(os.Stdout, agg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "输出错误: %v\n", err)
		os.Exit(1)
	}

	if agg.FailedLines > 0 {
		fmt.Fprintf(os.Stderr, "警告: %d / %d 行解析失败\n", agg.FailedLines, agg.TotalLines)
	}
}

func collectFiles(fileList []string, dirPath string, recursive bool) []string {
	var files []string

	for _, f := range fileList {
		info, err := os.Stat(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: 无法访问 %s: %v\n", f, err)
			continue
		}
		if info.IsDir() {
			dirFiles := collectLogFiles(f, recursive)
			files = append(files, dirFiles...)
		} else {
			files = append(files, f)
		}
	}

	if dirPath != "" {
		dirFiles := collectLogFiles(dirPath, recursive)
		files = append(files, dirFiles...)
	}

	return files
}

func collectLogFiles(dir string, recursive bool) []string {
	var files []string
	if recursive {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".log") {
				files = append(files, path)
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: 无法读取目录 %s: %v\n", dir, err)
			return files
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".log") {
				files = append(files, filepath.Join(dir, entry.Name()))
			}
		}
	}
	return files
}

func processFilesParallel(files []string, createParser func() (parser.Parser, error), fil *filter.Filter, agg *stats.Aggregator, showProgress bool) {
	if len(files) == 0 {
		return
	}

	workers := runtime.NumCPU()
	if workers > len(files) {
		workers = len(files)
	}

	fileChan := make(chan string, len(files))
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localP, err := createParser()
			if err != nil {
				fmt.Fprintf(os.Stderr, "警告: 创建解析器失败: %v\n", err)
				return
			}
			for path := range fileChan {
				var fileSize int64
				if showProgress {
					if info, err := os.Stat(path); err == nil {
						fileSize = info.Size()
					}
				}
				f, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "警告: 无法打开 %s: %v\n", path, err)
					continue
				}
				processReader(f, localP, fil, agg, path, showProgress, fileSize)
				f.Close()
			}
		}()
	}

	for _, f := range files {
		fileChan <- f
	}
	close(fileChan)

	wg.Wait()
}

func processReader(r io.Reader, p parser.Parser, fil *filter.Filter, agg *stats.Aggregator, name string, showProgress bool, totalSize int64) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var processed int64
	for scanner.Scan() {
		line := scanner.Text()
		processed += int64(len(line)) + 1

		agg.AddTotalLine()

		if strings.TrimSpace(line) == "" {
			continue
		}

		entry, err := p.Parse(line)
		if err != nil || entry == nil {
			agg.AddFailed()
			continue
		}

		if !fil.Match(entry) {
			continue
		}

		agg.Add(entry)

		if showProgress && totalSize > 100*1024*1024 && processed%(10*1024*1024) == 0 {
			percent := float64(processed) / float64(totalSize) * 100
			fmt.Fprintf(os.Stderr, "\r%s: %.1f%% (%d / %d MB)", name, percent, processed/(1024*1024), totalSize/(1024*1024))
		}
	}

	if showProgress && totalSize > 100*1024*1024 {
		fmt.Fprintf(os.Stderr, "\r%s: 100%% 完成\n", name)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 读取 %s 出错: %v\n", name, err)
	}
}
