package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

var stopwords = map[string]bool{
	"the": true, "and": true, "for": true, "are": true, "but": true, "not": true,
	"you": true, "all": true, "can": true, "had": true, "her": true, "was": true,
	"one": true, "our": true, "out": true, "has": true, "have": true, "been": true,
	"would": true, "could": true, "there": true, "their": true, "what": true,
	"about": true, "which": true, "when": true, "make": true, "like": true,
	"time": true, "just": true, "know": true, "take": true, "people": true,
	"into": true, "year": true, "your": true, "good": true, "some": true,
	"them": true, "than": true, "then": true, "now": true, "look": true,
	"only": true, "come": true, "its": true, "over": true, "think": true,
	"also": true, "back": true, "after": true, "use": true, "two": true,
	"how": true, "first": true, "well": true, "way": true, "even": true,
	"new": true, "want": true, "because": true, "any": true, "these": true,
	"give": true, "day": true, "most": true, "from": true, "with": true,
	"this": true, "that": true, "they": true, "will": true, "each": true,
	"made": true, "does": true, "other": true, "more": true, "such": true,
	"where": true, "here": true, "should": true, "being": true, "before": true,
	"true": true, "false": true, "null": true, "undefined": true, "function": true,
	"return": true, "var": true, "let": true, "const": true, "class": true,
	"export": true, "import": true, "default": true, "async": true, "await": true,
	"try": true, "catch": true, "finally": true, "throw": true,
	"if": true, "else": true, "switch": true, "case": true, "break": true,
	"continue": true, "while": true, "do": true, "in": true,
	"of": true, "typeof": true, "instanceof": true, "void": true, "delete": true,
}

var cssProperties = map[string]bool{
	"width": true, "height": true, "margin": true, "padding": true, "border": true,
	"color": true, "background": true, "font": true, "display": true, "position": true,
	"top": true, "bottom": true, "left": true, "right": true, "float": true,
	"clear": true, "overflow": true, "visibility": true, "opacity": true, "cursor": true,
	"transform": true, "transition": true, "animation": true, "flex": true, "grid": true,
	"align": true, "justify": true, "content": true, "items": true, "self": true,
	"order": true, "grow": true, "shrink": true, "basis": true, "wrap": true,
	"direction": true, "column": true, "row": true, "gap": true, "area": true,
	"template": true, "auto": true, "start": true, "end": true, "center": true,
	"space": true, "between": true, "around": true, "evenly": true, "stretch": true,
	"none": true, "block": true, "inline": true, "relative": true, "absolute": true,
	"fixed": true, "sticky": true, "static": true, "hidden": true, "visible": true,
	"scroll": true, "inherit": true, "initial": true, "unset": true, "revert": true,
	"size": true, "style": true, "weight": true, "family": true, "line": true,
	"text": true, "decoration": true, "shadow": true, "radius": true, "image": true,
	"repeat": true, "attachment": true, "clip": true, "origin": true, "blend": true,
	"filter": true, "outline": true, "resize": true, "appearance": true, "pointer": true,
	"events": true, "select": true, "touch": true, "action": true, "zoom": true,
	"min": true, "max": true, "em": true, "rem": true, "px": true, "vh": true, "vw": true,
	"solid": true, "dashed": true, "dotted": true, "double": true, "groove": true,
	"ridge": true, "inset": true, "outset": true, "collapse": true, "separate": true,
	"normal": true, "nowrap": true, "pre": true, "bold": true, "italic": true,
	"underline": true, "overline": true, "through": true, "blink": true, "capitalize": true,
	"uppercase": true, "lowercase": true, "rgb": true, "rgba": true, "hsl": true, "hsla": true,
	"url": true, "calc": true, "var": true, "attr": true, "counter": true, "rect": true,
}

var (
	jsonKeyRe    = regexp.MustCompile(`["']([a-zA-Z_][a-zA-Z0-9_]{2,})["']\s*:`)
	urlParamRe   = regexp.MustCompile(`[?&]([a-zA-Z_][a-zA-Z0-9_]{2,})=`)
	inputNameRe  = regexp.MustCompile(`(?i)name\s*=\s*["']([a-zA-Z_][a-zA-Z0-9_]{2,})["']`)
	jsVarRe      = regexp.MustCompile(`(?:var|let|const)\s+([a-zA-Z_][a-zA-Z0-9_]{2,})`)
	jsObjKeyRe   = regexp.MustCompile(`\.([a-zA-Z_][a-zA-Z0-9_]{2,})`)
	formActionRe = regexp.MustCompile(`(?i)action\s*=\s*["']([^"']+)["']`)
	dataAttrRe   = regexp.MustCompile(`data-([a-zA-Z][a-zA-Z0-9-]{2,})`)
	idClassRe    = regexp.MustCompile(`(?:id|class)\s*=\s*["']([a-zA-Z_][a-zA-Z0-9_-]{2,})["']`)
)

type WordCount struct {
	Word  string
	Count int
}

func main() {
	minLen := flag.Int("min", 3, "minimum word length")
	limit := flag.Int("limit", 0, "limit output to top N words (0 = no limit)")
	concurrency := flag.Int("c", 50, "number of concurrent requests")
	flag.Parse()

	wordFreq := make(map[string]int)
	var mu sync.Mutex

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(1),
	)
	c.Limit(&colly.LimitRule{
		Parallelism: *concurrency,
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		html, _ := e.DOM.Html()
		extractWords(html, &mu, wordFreq, *minLen)
	})

	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		if strings.HasSuffix(src, ".js") || strings.Contains(src, ".js?") {
			jsURL := e.Request.AbsoluteURL(src)
			fetchJS(jsURL, &mu, wordFreq, *minLen)
		}
	})

	c.OnHTML("script:not([src])", func(e *colly.HTMLElement) {
		extractWords(e.Text, &mu, wordFreq, *minLen)
	})

	c.OnResponse(func(r *colly.Response) {
		ct := r.Headers.Get("Content-Type")
		if strings.Contains(ct, "javascript") || strings.Contains(ct, "json") {
			extractWords(string(r.Body), &mu, wordFreq, *minLen)
		}
		extractWords(r.Request.URL.RawQuery, &mu, wordFreq, *minLen)
	})

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") {
			line = "https://" + line
		}
		if u, err := url.Parse(line); err == nil {
			extractWords(u.RawQuery, &mu, wordFreq, *minLen)
			extractWords(u.Path, &mu, wordFreq, *minLen)
		}
		c.Visit(line)
	}

	c.Wait()

	words := make([]WordCount, 0, len(wordFreq))
	for word, count := range wordFreq {
		words = append(words, WordCount{word, count})
	}

	sort.Slice(words, func(i, j int) bool {
		if words[i].Count == words[j].Count {
			return words[i].Word < words[j].Word
		}
		return words[i].Count > words[j].Count
	})

	outputLimit := len(words)
	if *limit > 0 && *limit < outputLimit {
		outputLimit = *limit
	}

	for i := 0; i < outputLimit; i++ {
		fmt.Println(words[i].Word)
	}
}

func fetchJS(jsURL string, mu *sync.Mutex, wordFreq map[string]int, minLen int) {
	c := colly.NewCollector()
	c.OnResponse(func(r *colly.Response) {
		extractWords(string(r.Body), mu, wordFreq, minLen)
	})
	c.Visit(jsURL)
}

func extractWords(content string, mu *sync.Mutex, wordFreq map[string]int, minLen int) {
	patterns := []*regexp.Regexp{
		jsonKeyRe,
		urlParamRe,
		inputNameRe,
		jsVarRe,
		jsObjKeyRe,
		dataAttrRe,
		idClassRe,
	}

	found := make(map[string]bool)

	for _, re := range patterns {
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				word := normalize(match[1])
				if isValidWord(word, minLen) {
					found[word] = true
				}
			}
		}
	}

	actionMatches := formActionRe.FindAllStringSubmatch(content, -1)
	for _, match := range actionMatches {
		if len(match) > 1 {
			if u, err := url.Parse(match[1]); err == nil {
				for key := range u.Query() {
					word := normalize(key)
					if isValidWord(word, minLen) {
						found[word] = true
					}
				}
			}
		}
	}

	mu.Lock()
	for word := range found {
		wordFreq[word]++
	}
	mu.Unlock()
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimPrefix(s, "-")
	s = strings.TrimSuffix(s, "-")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func isValidWord(word string, minLen int) bool {
	if len(word) < minLen {
		return false
	}
	if stopwords[word] {
		return false
	}
	if cssProperties[word] {
		return false
	}
	if isNumeric(word) {
		return false
	}
	return true
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
