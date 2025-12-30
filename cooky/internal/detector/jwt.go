package detector

import (
	"encoding/base64"
	"strings"
)

type JWTDetector struct{}

func (d *JWTDetector) Name() string { return "jwt" }

func (d *JWTDetector) Detect(value string) *Result {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return nil
	}

	if len(parts[0]) < 10 || len(parts[1]) < 10 || len(parts[2]) < 10 {
		return nil
	}

	for _, p := range parts[:2] {
		if !isBase64URLChars(p) {
			return nil
		}
	}

	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}

	headerStr := string(header)
	payloadStr := string(payload)

	if !strings.Contains(headerStr, `"alg"`) && !strings.Contains(headerStr, `"typ"`) {
		return nil
	}

	if !strings.HasPrefix(headerStr, "{") || !strings.HasSuffix(headerStr, "}") {
		return nil
	}

	if !strings.HasPrefix(payloadStr, "{") || !strings.HasSuffix(payloadStr, "}") {
		return nil
	}

	return &Result{
		Encoding: "jwt",
		Decoded:  headerStr + "." + payloadStr + "." + parts[2],
		Score:    0.99,
	}
}

func isBase64URLChars(s string) bool {
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
