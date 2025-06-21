package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Reset  = "\033[0m"
)

var (
	target      = flag.String("u", "", "Target URL")
	exclude     = flag.String("x", "", "Exclude methods (comma-separated)")
	include     = flag.String("i", "", "Include only these methods (comma-separated)")
	custom      = flag.String("c", "", "Add custom methods (comma-separated)")
	threads     = flag.Int("t", 10, "Number of threads")
	timeout     = flag.Int("timeout", 10, "Timeout in seconds")
	verbose     = flag.Bool("v", false, "Verbose output")
	silent      = flag.Bool("s", false, "Silent mode (only show working methods)")
)

func main() {
	showBanner()
	flag.Parse()

	var urls []string
	
	if *target != "" {
		urls = append(urls, *target)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			url := strings.TrimSpace(scanner.Text())
			if url != "" {
				urls = append(urls, url)
			}
		}
	}

	if len(urls) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No URLs provided. Use -u flag or pipe URLs via stdin\n")
		os.Exit(1)
	}

	methods := getHTTPMethods()
	methods = filterMethods(methods)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, *threads)
	
	for _, url := range urls {
		for _, method := range methods {
			wg.Add(1)
			go func(url, method string) {
				defer wg.Done()
				semaphore <- struct{}{}
				checkMethod(url, method)
				<-semaphore
			}(url, method)
		}
	}
	
	wg.Wait()
}

func showBanner() {
	fmt.Println(``)
}

func getHTTPMethods() []string {
	return []string{
		"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH",
		"TRACE", "CONNECT", "PROPFIND", "PROPPATCH", "MKCOL",
		"COPY", "MOVE", "LOCK", "UNLOCK", "VERSION-CONTROL",
		"REPORT", "CHECKOUT", "CHECKIN", "UNCHECKOUT", "MKWORKSPACE",
		"UPDATE", "LABEL", "MERGE", "BASELINE-CONTROL", "MKACTIVITY",
		"ORDERPATCH", "ACL", "SEARCH", "PURGE", "LINK", "UNLINK",
		"VIEW", "WRAPPED", "Extension-mothed", "REBIND", "UNBIND",
		"BIND", "MSEARCH", "NOTIFY", "SUBSCRIBE", "UNSUBSCRIBE",
		"DEBUG", "TRACK", "MKCALENDAR", "MKREDIRECTREF", "UPDATEREDIRECTREF",
	}
}

func filterMethods(methods []string) []string {
	var filtered []string
	excludeMap := make(map[string]bool)
	includeMap := make(map[string]bool)

	if *exclude != "" {
		for _, method := range strings.Split(*exclude, ",") {
			excludeMap[strings.TrimSpace(strings.ToUpper(method))] = true
		}
	}

	if *include != "" {
		for _, method := range strings.Split(*include, ",") {
			includeMap[strings.TrimSpace(strings.ToUpper(method))] = true
		}
	}

	if *custom != "" {
		for _, method := range strings.Split(*custom, ",") {
			methods = append(methods, strings.TrimSpace(strings.ToUpper(method)))
		}
	}

	for _, method := range methods {
		method = strings.ToUpper(method)
		if excludeMap[method] {
			continue
		}
		if len(includeMap) > 0 && !includeMap[method] {
			continue
		}
		filtered = append(filtered, method)
	}

	return filtered
}

func checkMethod(url, method string) {
	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		if *verbose {
			fmt.Printf("%s[ERROR]%s %s - %s: %v\n", Red, Reset, method, url, err)
		}
		return
	}

	req.Header.Set("User-Agent", "meth/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		if *verbose {
			fmt.Printf("%s[TIMEOUT]%s %s - %s\n", Yellow, Reset, method, url)
		}
		return
	}
	defer resp.Body.Close()

	if isMethodAllowed(resp.StatusCode, method, resp.Header.Get("Allow")) {
		statusColor := getStatusColor(resp.StatusCode)
		methodColor := getMethodColor(method)
		fmt.Printf("%s[%s%d%s%s]%s%s[%s%s%s%s]%s %s\n", 
			White, statusColor, resp.StatusCode, Reset, White, Reset,
			White, methodColor, method, Reset, White, Reset, url)
	} else if *verbose {
		statusColor := getStatusColor(resp.StatusCode)
		fmt.Printf("%s[%s%d%s%s]%s %s - %s\n", White, statusColor, resp.StatusCode, Reset, White, Reset, method, url)
	}
}

func isMethodAllowed(statusCode int, method string, allowHeader string) bool {
	if statusCode == 405 {
		return false
	}
	
	if statusCode == 501 {
		return false
	}
	
	if statusCode == 404 {
		return false
	}
	
	if statusCode == 400 && (method == "TRACE" || method == "TRACK" || method == "DEBUG") {
		return false
	}
	
	if statusCode >= 500 && statusCode < 600 {
		return false
	}
	
	if allowHeader != "" && method == "OPTIONS" {
		allowedMethods := strings.Split(strings.ToUpper(allowHeader), ",")
		for i := range allowedMethods {
			allowedMethods[i] = strings.TrimSpace(allowedMethods[i])
		}
		for _, allowed := range allowedMethods {
			if allowed == method {
				return true
			}
		}
	}
	
	if statusCode >= 200 && statusCode < 300 {
		return true
	}
	
	if statusCode >= 300 && statusCode < 400 {
		return method == "GET" || method == "HEAD" || method == "OPTIONS"
	}
	
	if statusCode == 401 || statusCode == 403 {
		return true
	}
	
	return false
}

func getMethodColor(method string) string {
	switch method {
	case "DELETE", "PURGE":
		return Red
	case "PUT", "PATCH", "POST":
		return Yellow
	case "TRACE", "TRACK", "DEBUG":
		return Purple
	case "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK":
		return Cyan
	case "OPTIONS", "HEAD":
		return Blue
	default:
		return Green
	}
}

func getStatusColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return Green
	case statusCode >= 300 && statusCode < 400:
		return Blue
	case statusCode == 401 || statusCode == 403:
		return Yellow
	case statusCode >= 400 && statusCode < 500:
		return Purple
	case statusCode >= 500:
		return Red
	default:
		return White
	}
}
