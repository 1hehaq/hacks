package detector

import (
	"net/url"
	"strings"
)

type URLDetector struct{}

func (d *URLDetector) Name() string { return "url" }

func (d *URLDetector) Detect(value string) *Result {
	if !strings.Contains(value, "%") {
		return nil
	}

	percentCount := strings.Count(value, "%")
	if percentCount < 1 {
		return nil
	}

	decoded, err := url.QueryUnescape(value)
	if err != nil {
		return nil
	}

	if decoded == value || !hasMeaningfulChange(value, decoded) {
		return nil
	}

	encodedChars := 0
	for i := 0; i < len(value)-2; i++ {
		if value[i] == '%' {
			c1, c2 := value[i+1], value[i+2]
			if isHexChar(c1) && isHexChar(c2) {
				encodedChars++
			}
		}
	}

	if encodedChars == 0 {
		return nil
	}

	ratio := float64(encodedChars) / float64(len(value))
	score := 0.7 + (ratio * 0.29)
	if score > 0.99 {
		score = 0.99
	}

	return &Result{
		Encoding: "url",
		Decoded:  decoded,
		Score:    score,
	}
}

func isHexChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func hasMeaningfulChange(original, decoded string) bool {
	if len(original) == len(decoded) {
		diff := 0
		for i := 0; i < len(original); i++ {
			if original[i] != decoded[i] {
				diff++
			}
		}
		return diff > 0
	}
	return len(original) != len(decoded)
}
