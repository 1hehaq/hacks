package fetcher

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/1hehaq/cooky/internal/analyzer"
)

type Result struct {
	URL     string            `json:"url"`
	Cookies []analyzer.Cookie `json:"cookies,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type CookieResult struct {
	URL    string
	Cookie analyzer.Cookie
	Error  string
}

type Fetcher struct {
	client      *http.Client
	userAgent   string
	concurrency int
}

func New(timeout time.Duration, concurrency int) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		userAgent:   "cooky/1.0",
		concurrency: concurrency,
	}
}

func (f *Fetcher) Fetch(ctx context.Context, rawURL string) Result {
	result := Result{URL: rawURL}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
		result.URL = rawURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Header.Values("Set-Cookie") {
		name, value := parseCookie(cookie)
		if name != "" {
			result.Cookies = append(result.Cookies, analyzer.Analyze(name, value))
		}
	}

	return result
}

func (f *Fetcher) FetchStream(ctx context.Context, urls []string) <-chan CookieResult {
	out := make(chan CookieResult)
	sem := make(chan struct{}, f.concurrency)
	var wg sync.WaitGroup

	for _, u := range urls {
		wg.Add(1)
		go func(rawURL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			url := rawURL
			if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
				url = "https://" + rawURL
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				out <- CookieResult{URL: url, Error: err.Error()}
				return
			}
			req.Header.Set("User-Agent", f.userAgent)

			resp, err := f.client.Do(req)
			if err != nil {
				out <- CookieResult{URL: url, Error: err.Error()}
				return
			}
			defer resp.Body.Close()

			for _, cookie := range resp.Header.Values("Set-Cookie") {
				name, value := parseCookie(cookie)
				if name != "" {
					c := analyzer.Analyze(name, value)
					if c.Encoding != "" && c.Decoded != "" && c.Value != c.Decoded {
						out <- CookieResult{URL: url, Cookie: c}
					}
				}
			}
		}(u)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (f *Fetcher) FetchMany(ctx context.Context, urls []string) []Result {
	results := make([]Result, len(urls))
	sem := make(chan struct{}, f.concurrency)
	var wg sync.WaitGroup

	for i, u := range urls {
		wg.Add(1)
		go func(idx int, rawURL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = f.Fetch(ctx, rawURL)
		}(i, u)
	}

	wg.Wait()
	return results
}

func parseCookie(raw string) (name, value string) {
	parts := strings.SplitN(raw, ";", 2)
	if len(parts) == 0 {
		return "", ""
	}
	kv := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
	if len(kv) >= 1 {
		name = kv[0]
	}
	if len(kv) >= 2 {
		value = kv[1]
	}
	return name, value
}
