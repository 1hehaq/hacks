package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var blacklist = map[string]bool{
	"await": true, "break": true, "case": true, "catch": true, "class": true, "const": true, "continue": true, "debugger": true, "default": true, "delete": true, "do": true, "else": true, "enum": true, "export": true, "extends": true, "false": true, "finally": true, "for": true, "function": true, "if": true, "implements": true, "import": true, "in": true, "instanceof": true, "interface": true, "let": true, "new": true, "null": true, "package": true, "private": true, "protected": true, "public": true, "return": true, "super": true, "switch": true, "static": true, "this": true, "throw": true, "try": true, "true": true, "typeof": true, "var": true, "void": true, "while": true, "with": true, "abstract": true, "boolean": true, "byte": true, "char": true, "double": true, "final": true, "float": true, "goto": true, "int": true, "long": true, "native": true, "short": true, "synchronized": true, "throws": true, "transient": true, "volatile": true, "window": true, "document": true, "undefined": true, "eval": true, "alert": true, "console": true,
}

func main() {
	js := flag.Bool("js", false, "extract js keywords (supported only for .js files)")
	url := flag.Bool("url", false, "extract path/parameter from urls")
	flag.Parse()
	
	scanner := bufio.NewScanner(os.Stdin)
	if flag.NArg() > 0 {
		for _, arg := range flag.Args() {
			process(arg, *js, *url)
		}
	} else {
		for scanner.Scan() {
			if line := strings.TrimSpace(scanner.Text()); line != "" {
				process(line, *js, *url)
			}
		}
	}
}

func process(input string, js, url bool) {
	if js {
		if resp, err := http.Get(input); err == nil && resp.StatusCode == 200 {
			if body, err := io.ReadAll(resp.Body); err == nil {
				extract(string(body))
			}
			resp.Body.Close()
		}
	} else if url {
		extractURL(input)
	}
}

func extract(content string) {
	words := make(map[string]bool)
	re := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)
	for _, word := range re.FindAllString(content, -1) {
		if len(word) > 1 && !blacklist[strings.ToLower(word)] {
			words[word] = true
		}
	}
	for word := range words {
		fmt.Println(word)
	}
}

func extractURL(urlStr string) {
	if u, err := url.Parse(urlStr); err == nil {
		words := make(map[string]bool)
		text := u.Path + "?" + u.RawQuery
		for _, part := range regexp.MustCompile(`[^/\?&=\-_.]+`).FindAllString(text, -1) {
			if len(part) > 0 && !blacklist[strings.ToLower(part)] {
				words[part] = true
				for _, camel := range splitCamel(part) {
					if len(camel) > 0 && !blacklist[strings.ToLower(camel)] {
						words[camel] = true
					}
				}
			}
		}
		for word := range words {
			fmt.Println(word)
		}
	}
}

func splitCamel(s string) []string {
	var result []string
	var current strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && current.Len() > 0 {
			result = append(result, current.String())
			current.Reset()
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
