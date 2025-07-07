package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	domain := getDomain()
	if domain == "" {
		log.Fatal("no domain provided")
	}

	domain = normalizeURL(domain)
	content := fetch(domain)
	if content == nil {
		log.Fatal("failed to fetch domain")
	}

	baseDir := sanitizeDomain(domain)
	if err := os.MkdirAll(filepath.Join(baseDir, "js"), 0755); err != nil {
		log.Fatal(err)
	}

	jsFiles := gatherJS(domain, content)
	seen := make(map[string]bool)

	for _, js := range jsFiles {
		if seen[js] {
			continue
		}
		seen[js] = true

		content := fetch(js)
		if content == nil {
			continue
		}

		saveJS(baseDir, js, content)
		for _, e := range filterEndpoints(findEndpoints(string(content))) {
			fmt.Println(e)
		}
	}
}

func getDomain() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func normalizeURL(u string) string {
	if !strings.HasPrefix(u, "http") {
		return "https://" + u
	}
	return u
}

func sanitizeDomain(d string) string {
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimRight(d, "/")
	return strings.Split(d, ":")[0]
}

func gatherJS(u string, content []byte) []string {
	re := regexp.MustCompile(`<script[^>]*src\s*=\s*['"](.*?)['"]`)
	matches := re.FindAllSubmatch(content, -1)
	jsFiles := make([]string, 0)
	base, _ := url.Parse(u)
	for _, m := range matches {
		rel, err := url.Parse(string(m[1]))
		if err != nil {
			continue
		}
		jsFiles = append(jsFiles, base.ResolveReference(rel).String())
	}
	return jsFiles
}

func fetch(u string) []byte {
	resp, err := http.Get(u)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	return body
}

func saveJS(base, urlStr string, content []byte) {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Println(err)
		return
	}
	fname := strings.TrimPrefix(u.Path, "/")
	fname = strings.ReplaceAll(fname, "/", "_")
	if fname == "" {
		fname = "index.js"
	}
	path := filepath.Join(base, "js", fname)
	if err := os.WriteFile(path, content, 0644); err != nil {
		log.Println(err)
	}
}

func findEndpoints(content string) []string {
	re := regexp.MustCompile(`(https?://[^\s'"]+)|['"](/[^'"\s]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	endpoints := make([]string, 0)
	for _, m := range matches {
		if m[1] != "" {
			endpoints = append(endpoints, m[1])
		} else if len(m) > 2 && m[2] != "" {
			endpoints = append(endpoints, m[2])
		}
	}
	return endpoints
}

func filterEndpoints(endpoints []string) []string {
	unique := make(map[string]bool)
	var filtered []string
	for _, e := range endpoints {
		if !unique[e] && len(e) > 3 && !strings.Contains(e, " ") {
			unique[e] = true
			filtered = append(filtered, e)
		}
	}
	return filtered
}
