package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	base := "xls|xml|xlsx|json|pdf|sql|doc|docx|pptx|txt|zip|tar\\.gz|tgz|bak|7z|rar|log|cache|secret|db|backup|yml|gz|config|csv|yaml|md|md5|exe|dll|bin|ini|bat|sh|tar|deb|rpm|iso|img|apk|msi|dmg|tmp|crt|pem|key|pub|asc"
	target := flag.String("t", "", "Single target domain")
	include := flag.String("in", "", "Extensions to include (comma separated)")
	exclude := flag.String("ex", "", "Extensions to exclude (comma separated)")
	flag.Parse()

	exts := base
	if *include != "" {
		for _, ext := range strings.Split(*include, ",") {
			escaped := strings.Replace(ext, ".", "\\.", -1)
			exts = exts + "|" + escaped
		}
	}
	if *exclude != "" {
		removeSet := make(map[string]bool)
		for _, ext := range strings.Split(*exclude, ",") {
			escaped := strings.Replace(ext, ".", "\\.", -1)
			removeSet[escaped] = true
		}
		var newExts []string
		for _, ext := range strings.Split(exts, "|") {
			if !removeSet[ext] {
				newExts = append(newExts, ext)
			}
		}
		exts = strings.Join(newExts, "|")
	}

	domains := make(chan string)
	var wg sync.WaitGroup

	go func() {
		switch {
		case *target != "":
			domains <- strings.TrimSpace(*target)
		case len(flag.Args()) > 0:
			for _, domain := range flag.Args() {
				if trimmed := strings.TrimSpace(domain); trimmed != "" {
					domains <- trimmed
				}
			}
		default:
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				if trimmed := strings.TrimSpace(scanner.Text()); trimmed != "" {
					domains <- trimmed
				}
			}
		}
		close(domains)
	}()

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 15 * time.Second,
		},
	}

	sem := make(chan struct{}, 10)
	ctx := context.Background()

	for domain := range domains {
		wg.Add(1)
		sem <- struct{}{}
		go func(d string) {
			defer wg.Done()
			defer func() { <-sem }()

			cleanDomain := strings.TrimPrefix(d, "*.")
			wildcardDomain := "*." + cleanDomain
			
			params := url.Values{}
			params.Set("url", wildcardDomain+"/*")
			params.Set("collapse", "urlkey")
			params.Set("output", "text")
			params.Set("fl", "original")
			params.Set("filter", "original:.*\\."+"("+exts+")$")
			
			req, _ := http.NewRequestWithContext(ctx, "GET", "https://web.archive.org/cdx/search/cdx?"+params.Encode(), nil)
			req.Header.Set("User-Agent", "arqx/1.0")
			
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s: %v\n", d, err)
				return
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != 200 {
				fmt.Fprintf(os.Stderr, "error: %s: status %d\n", d, resp.StatusCode)
				return
			}
			
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}(domain)
	}

	wg.Wait()
}
