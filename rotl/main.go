package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			rotate(line)
		}
	}
}

func rotate(u string) {
	parsed, err := url.Parse(u)
	if err != nil {
		return
	}
	
	path := parsed.Path
	if path == "" {
		path = "/"
	}
	
	base := parsed.Scheme + "://" + parsed.Host
	query := ""
	if parsed.RawQuery != "" {
		query = "?" + parsed.RawQuery
	}
	
	patterns := []string{
		path + "/", path + ".example", path + ".sample", path + ".template",
		"..%3B" + path + "/", path + "..%2f", path + "%09", path + "%23",
		path + "..%00", path + ";%09", path + ";%09..", path + ";%09..;", path + ";%2f..",
		"." + path, "%0A" + path, "%0D%0A" + path, "%0D" + path, "%2e" + path + "/",
		path + "%20", path + "%2520", "%u002e%u002e/%u002e%u002e" + path,
		"%2e%2e%2f" + path + "/", "%2E" + path, path + ".old", path + "?.css", path + "?.js",
		"_" + path, path + "_", "_" + path + "_", "..;" + path + "/", "..;/..;" + path + "/", ".." + path,
		"-" + path, "~" + path, path + "..;/", path + ";/", path + "#", path + "/~",
		"!" + path, "#" + path + "/", "-" + path + "/", path + "~", path + "/.git/config",
		path + "/.env", path + ".", path + "/*", path + "/?",
		"//" + strings.TrimPrefix(path, "/"), path + "//", strings.ToUpper(path), strings.ToLower(path),
		path + "%2f", path + "%252f", path + "%5c", path + "%255c", path + "%00", path + "%2500",
		path + ".json", path + ".xml", path + ".html", path + ".txt", path + ".bak",
		path + "%20/", path + "/.", path + "/..", path + "/./", path + "/../",
	}
	
	seen := make(map[string]bool)
	seen[u] = true
	
	for _, p := range patterns {
		full := base + p + query
		if !seen[full] {
			fmt.Println(full)
			seen[full] = true
		}
	}
}
