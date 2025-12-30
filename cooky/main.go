package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/1hehaq/cooky/internal/fetcher"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var (
	tagStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	arrowStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#04B575"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	addStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	equalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

func main() {
	urlFlag := flag.String("u", "", "URL to fetch cookies from")
	timeout := flag.Duration("t", 10*time.Second, "HTTP timeout")
	concurrency := flag.Int("c", 10, "Concurrent requests")
	jsonFlag := flag.Bool("json", false, "Output raw JSON format")
	flag.Parse()

	urls := collectURLs(*urlFlag)
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "usage: cooky -u <url> or pipe URLs via stdin")
		os.Exit(1)
	}

	f := fetcher.New(*timeout, *concurrency)

	if *jsonFlag {
		results := f.FetchMany(context.Background(), urls)
		enc := json.NewEncoder(os.Stdout)
		for _, r := range results {
			enc.Encode(r)
		}
	} else {
		stream := f.FetchStream(context.Background(), urls)
		dmp := diffmatchpatch.New()
		
		for result := range stream {
			if result.Error != "" {
				fmt.Fprintln(os.Stderr, errorStyle.Render(result.URL+": "+result.Error))
				continue
			}
			
			c := result.Cookie
			tag := tagStyle.Render("[" + c.Encoding + "]")
			arrow := arrowStyle.Render("→")
			coloredDiff := renderColoredDiff(dmp, c.Value, c.Decoded)
			fmt.Printf("%s %s %s %s\n", tag, c.Value, arrow, coloredDiff)
		}
	}
}

func renderColoredDiff(dmp *diffmatchpatch.DiffMatchPatch, original, decoded string) string {
	diffs := dmp.DiffMain(original, decoded, false)
	diffs = dmp.DiffCleanupSemantic(diffs)
	
	var result strings.Builder
	for _, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			result.WriteString(addStyle.Render(d.Text))
		case diffmatchpatch.DiffEqual:
			result.WriteString(equalStyle.Render(d.Text))
		}
	}
	return result.String()
}

func collectURLs(urlFlag string) []string {
	var urls []string
	if urlFlag != "" {
		urls = append(urls, urlFlag)
	}
	if !isTerminal() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if line := strings.TrimSpace(scanner.Text()); line != "" && !strings.HasPrefix(line, "#") {
				urls = append(urls, line)
			}
		}
	}
	return urls
}

func isTerminal() bool {
	fi, _ := os.Stdin.Stat()
	return fi.Mode()&os.ModeCharDevice != 0
}
