package detector

import (
	"encoding/hex"
	"math"
	"strings"
	"unicode"
)

type HexDetector struct{}

func (d *HexDetector) Name() string { return "hex" }

func (d *HexDetector) Detect(value string) *Result {
	if len(value) < 16 {
		return nil
	}

	if len(value)%2 != 0 {
		return nil
	}

	if !isHexChars(value) {
		return nil
	}

	hasLower := strings.ContainsAny(value, "abcdef")
	hasUpper := strings.ContainsAny(value, "ABCDEF")
	if hasLower && hasUpper {
		return nil
	}

	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil
	}

	if !isPrintableHex(decoded) {
		return nil
	}

	entropy := calculateHexEntropy(value)
	if entropy < 3.0 {
		return nil
	}

	score := calculateHexScore(value, decoded, entropy)
	if score < 0.5 {
		return nil
	}

	return &Result{
		Encoding: "hex",
		Decoded:  string(decoded),
		Score:    score,
	}
}

func isHexChars(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func isPrintableHex(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	printable := 0
	for _, b := range data {
		r := rune(b)
		if unicode.IsPrint(r) || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
	}

	return float64(printable)/float64(len(data)) > 0.95
}

func calculateHexEntropy(s string) float64 {
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}

	var entropy float64
	length := float64(len(s))
	for _, count := range freq {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func calculateHexScore(original string, decoded []byte, entropy float64) float64 {
	score := 0.4

	if len(original) >= 32 {
		score += 0.2
	}

	if entropy > 3.5 {
		score += 0.15
	}

	s := string(decoded)
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		score += 0.25
	} else if strings.Contains(s, "=") || strings.Contains(s, ":") {
		score += 0.15
	}

	if score > 0.99 {
		score = 0.99
	}

	return score
}
