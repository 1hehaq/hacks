package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type HTTPXResult struct {
	URL           string `json:"url"`
	Host          string `json:"host"`
	StatusCode    int    `json:"status_code"`
	ContentLength int64  `json:"content_length"`
	ContentType   string `json:"content_type"`
	Title         string `json:"title"`
	ServerHeader  string `json:"webserver"`
	Body          string `json:"body"`
}

type Fingerprint struct {
	URL           string
	Host          string
	StatusCode    int
	ServerHeader  string
	ContentType   string
	ContentLength int64
	Title         string
	StructureHash string
}

type Group struct {
	Key     string
	Count   int
	Members []*Fingerprint
}

type Deviant struct {
	Target   *Fingerprint
	Reason   string
	Baseline string
}

var (
	threshold   float64
	timeout     int
	httpxMode   bool
	verbose     bool
	tagPattern  = regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9]*)[^>]*>`)
)

func main() {
	flag.Float64Var(&threshold, "threshold", 5.0, "Percentage threshold for anomaly detection (default: 5%)")
	flag.IntVar(&timeout, "timeout", 10, "HTTP request timeout in seconds")
	flag.BoolVar(&httpxMode, "httpx", false, "Parse httpx JSON output from stdin")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.Parse()

	fingerprints := collectFingerprints()
	if len(fingerprints) == 0 {
		fmt.Fprintln(os.Stderr, "[!] No targets processed")
		os.Exit(1)
	}

	deviants := analyze(fingerprints)
	printDeviants(deviants)
}

func collectFingerprints() []*Fingerprint {
	var fingerprints []*Fingerprint
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var fp *Fingerprint
		if httpxMode {
			fp = parseHTTPX(line)
		} else {
			fp = probeURL(line)
		}

		if fp != nil {
			fingerprints = append(fingerprints, fp)
			if verbose {
				fmt.Fprintf(os.Stderr, "[*] Fingerprinted: %s\n", fp.Host)
			}
		}
	}

	return fingerprints
}

func parseHTTPX(line string) *Fingerprint {
	var result HTTPXResult
	if err := json.Unmarshal([]byte(line), &result); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "[!] Failed to parse JSON: %s\n", err)
		}
		return nil
	}

	host := result.Host
	if host == "" {
		host = extractHost(result.URL)
	}

	structHash := computeStructureHash(result.Body)

	return &Fingerprint{
		URL:           result.URL,
		Host:          host,
		StatusCode:    result.StatusCode,
		ServerHeader:  normalizeServer(result.ServerHeader),
		ContentType:   result.ContentType,
		ContentLength: result.ContentLength,
		Title:         result.Title,
		StructureHash: structHash,
	}
}

func probeURL(url string) *Fingerprint {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "[!] Invalid URL %s: %s\n", url, err)
		}
		return nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; deviant/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "[!] Request failed for %s: %s\n", url, err)
		}
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	bodyStr := string(body)

	return &Fingerprint{
		URL:           url,
		Host:          extractHost(url),
		StatusCode:    resp.StatusCode,
		ServerHeader:  normalizeServer(resp.Header.Get("Server")),
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		Title:         extractTitle(bodyStr),
		StructureHash: computeStructureHash(bodyStr),
	}
}

func computeStructureHash(body string) string {
	if body == "" {
		return "empty"
	}

	matches := tagPattern.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return "notags"
	}

	var tags []string
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, strings.ToLower(match[1]))
		}
	}

	structure := strings.Join(tags, ">")
	hash := md5.Sum([]byte(structure))
	return hex.EncodeToString(hash[:8])
}

func extractTitle(body string) string {
	titleRe := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	match := titleRe.FindStringSubmatch(body)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func extractHost(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}
	if idx := strings.Index(url, ":"); idx != -1 {
		url = url[:idx]
	}
	return url
}

func normalizeServer(server string) string {
	server = strings.TrimSpace(server)
	if server == "" {
		return "unknown"
	}
	return server
}

func analyze(fingerprints []*Fingerprint) []Deviant {
	total := len(fingerprints)
	thresholdCount := int(float64(total) * (threshold / 100.0))
	if thresholdCount < 1 {
		thresholdCount = 1
	}

	structGroups := groupBy(fingerprints, func(fp *Fingerprint) string {
		return fp.StructureHash
	})

	serverGroups := groupBy(fingerprints, func(fp *Fingerprint) string {
		return fp.ServerHeader
	})

	baselineStruct := findBaseline(structGroups)
	baselineServer := findBaseline(serverGroups)

	var deviants []Deviant

	for _, fp := range fingerprints {
		var reasons []string

		structGroup := structGroups[fp.StructureHash]
		if structGroup.Count <= thresholdCount && fp.StructureHash != baselineStruct.Key {
			reasons = append(reasons, fmt.Sprintf("Unique Structure (hash: %s)", fp.StructureHash[:8]))
		}

		if fp.ServerHeader != baselineServer.Key {
			serverGroup := serverGroups[fp.ServerHeader]
			if serverGroup.Count <= thresholdCount {
				reasons = append(reasons, fmt.Sprintf("Unique Server '%s' - Baseline is '%s'", fp.ServerHeader, baselineServer.Key))
			} else {
				reasons = append(reasons, fmt.Sprintf("Tech Drift: '%s' vs baseline '%s'", fp.ServerHeader, baselineServer.Key))
			}
		}

		if len(reasons) > 0 {
			deviants = append(deviants, Deviant{
				Target:   fp,
				Reason:   strings.Join(reasons, "; "),
				Baseline: baselineServer.Key,
			})
		}
	}

	sort.Slice(deviants, func(i, j int) bool {
		return deviants[i].Target.Host < deviants[j].Target.Host
	})

	return deviants
}

func groupBy(fingerprints []*Fingerprint, keyFunc func(*Fingerprint) string) map[string]*Group {
	groups := make(map[string]*Group)
	for _, fp := range fingerprints {
		key := keyFunc(fp)
		if groups[key] == nil {
			groups[key] = &Group{Key: key}
		}
		groups[key].Count++
		groups[key].Members = append(groups[key].Members, fp)
	}
	return groups
}

func findBaseline(groups map[string]*Group) *Group {
	var baseline *Group
	for _, g := range groups {
		if baseline == nil || g.Count > baseline.Count {
			baseline = g
		}
	}
	return baseline
}

func printDeviants(deviants []Deviant) {
	if len(deviants) == 0 {
		fmt.Fprintln(os.Stderr, "[*] No deviants detected. All targets conform to baseline.")
		return
	}

	for _, d := range deviants {
		fmt.Printf("[DEVIANT] %s (Reason: %s)\n", d.Target.Host, d.Reason)
	}

	fmt.Fprintf(os.Stderr, "\n[*] Analysis complete: %d deviant(s) identified\n", len(deviants))
}
