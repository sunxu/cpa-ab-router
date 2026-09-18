package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sunxu/cpa-ab-router/internal/classifier"
)

type row struct {
	File           string
	Route          string
	Classification string
	Reason         string
	Score          int
	Segments       int
}

func main() {
	out := flag.String("o", "router-audit", "output directory")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: cpa-ab-audit [-o DIR] LOG_DIR_OR_FILE")
		os.Exit(2)
	}
	files, err := collect(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	rows := make([]row, 0, len(files))
	failed := 0
	for _, path := range files {
		body, err := extractRequestBody(path)
		if err != nil {
			failed++
			rows = append(rows, row{File: path, Route: "B", Classification: "uncertain", Reason: "log_parse_failed"})
			continue
		}
		r := classifier.ClassifyJSON(body)
		rows = append(rows, row{path, r.Route, r.Classification, r.Reason, r.Score, r.Segments})
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fatal(err)
	}
	if err := writeCSV(filepath.Join(*out, "requests.csv"), rows); err != nil {
		fatal(err)
	}

	reasons := map[string]int{}
	a, b := 0, 0
	for _, r := range rows {
		if r.Route == "A" {
			a++
		} else {
			b++
		}
		reasons[r.Reason]++
	}
	summary := map[string]any{
		"files_seen": len(files), "parsed_or_fail_closed": len(rows), "log_parse_failed": failed,
		"route_A": a, "route_B": b, "A_ratio_pct": pct(a, len(rows)), "B_ratio_pct": pct(b, len(rows)),
		"reasons": reasons,
	}
	data, _ := json.MarshalIndent(summary, "", "  ")
	if err := os.WriteFile(filepath.Join(*out, "summary.json"), data, 0o644); err != nil {
		fatal(err)
	}
	fmt.Println(string(data))
}

func collect(path string) ([]string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return []string{path}, nil
	}
	var out []string
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".log") {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

func extractRequestBody(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	s.Buffer(buf, 16*1024*1024)
	found := false
	var b strings.Builder
	for s.Scan() {
		line := s.Text()
		if !found {
			if line == "=== REQUEST BODY ===" {
				found = true
			}
			continue
		}
		if strings.HasPrefix(line, "=== API REQUEST ") && strings.HasSuffix(line, " ===") {
			break
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if !found || strings.TrimSpace(b.String()) == "" {
		return nil, fmt.Errorf("request body not found")
	}
	return []byte(strings.TrimSpace(b.String())), nil
}

func writeCSV(path string, rows []row) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	_ = w.Write([]string{"file", "route", "classification", "reason", "score", "segments"})
	for _, r := range rows {
		_ = w.Write([]string{r.File, r.Route, r.Classification, r.Reason, fmt.Sprint(r.Score), fmt.Sprint(r.Segments)})
	}
	return w.Error()
}
func pct(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) * 100 / float64(d)
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
