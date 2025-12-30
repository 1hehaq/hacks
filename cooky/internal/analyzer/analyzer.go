package analyzer

import "github.com/1hehaq/cooky/internal/detector"

type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Encoding string  `json:"encoding,omitempty"`
	Decoded  string  `json:"decoded,omitempty"`
	Score    float64 `json:"score,omitempty"`
}

func Analyze(name, value string) Cookie {
	c := Cookie{
		Name:  name,
		Value: value,
	}

	if result := detector.Detect(value); result != nil {
		c.Encoding = result.Encoding
		c.Decoded = result.Decoded
		c.Score = result.Score
	}

	return c
}
