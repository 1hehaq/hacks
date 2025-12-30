package detector

type Result struct {
	Encoding string
	Decoded  string
	Score    float64
}

type Detector interface {
	Name() string
	Detect(value string) *Result
}

var detectors = []Detector{
	&JWTDetector{},
	&URLDetector{},
	&Base64Detector{},
	&HexDetector{},
}

func Detect(value string) *Result {
	var best *Result
	for _, d := range detectors {
		if r := d.Detect(value); r != nil {
			if best == nil || r.Score > best.Score {
				best = r
			}
		}
	}
	return best
}
