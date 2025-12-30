package detector

import (
	"encoding/base64"
	"math"
	"strings"
	"unicode"
)

type Base64Detector struct{}

func (d *Base64Detector) Name() string { return "base64" }

func (d *Base64Detector) Detect(value string) *Result {
	if len(value) < 8 {
		return nil
	}

	if !isBase64Chars(value) {
		return nil
	}

	var decoded []byte
	var err error
	var encoding string

	paddedLen := len(value)
	if paddedLen%4 != 0 {
		paddedLen += 4 - (paddedLen % 4)
	}
	padded := value + strings.Repeat("=", paddedLen-len(value))

	if decoded, err = base64.StdEncoding.DecodeString(padded); err == nil {
		encoding = "base64"
	} else if decoded, err = base64.URLEncoding.DecodeString(padded); err == nil {
		encoding = "base64url"
	} else if decoded, err = base64.RawStdEncoding.DecodeString(value); err == nil {
		encoding = "base64"
	} else if decoded, err = base64.RawURLEncoding.DecodeString(value); err == nil {
		encoding = "base64url"
	}

	if len(decoded) == 0 {
		return nil
	}

	if !isPrintableWithStructure(decoded) {
		return nil
	}

	entropy := calculateEntropy(value)
	if entropy < 3.5 {
		return nil
	}

	score := calculateBase64Score(value, decoded, entropy)
	if score < 0.5 {
		return nil
	}

	return &Result{
		Encoding: encoding,
		Decoded:  string(decoded),
		Score:    score,
	}
}

func isBase64Chars(s string) bool {
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func isPrintableWithStructure(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	printable := 0
	alphanumeric := 0
	for _, b := range data {
		r := rune(b)
		if unicode.IsPrint(r) || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') {
			alphanumeric++
		}
	}

	printableRatio := float64(printable) / float64(len(data))
	if printableRatio < 0.9 {
		return false
	}

	s := string(data)
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") ||
		strings.Contains(s, "=") || strings.Contains(s, "&") ||
		strings.Contains(s, ":") || strings.Contains(s, "|") {
		return true
	}

	alphaRatio := float64(alphanumeric) / float64(len(data))
	return alphaRatio > 0.3
}

func calculateEntropy(s string) float64 {
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

func calculateBase64Score(original string, decoded []byte, entropy float64) float64 {
	score := 0.5

	if len(original) >= 20 {
		score += 0.15
	}

	if entropy > 4.5 {
		score += 0.15
	} else if entropy > 4.0 {
		score += 0.10
	}

	s := string(decoded)
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		score += 0.2
	} else if strings.Contains(s, "=") && strings.Contains(s, "&") {
		score += 0.15
	} else if strings.Contains(s, ":") {
		score += 0.1
	}

	if score > 0.99 {
		score = 0.99
	}

	return score
}
