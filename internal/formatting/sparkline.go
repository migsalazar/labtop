package formatting

import (
	"math"
	"strings"
)

var sparklineLevels = [...]rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// SparklinePercent renders newest percentage samples at a fixed 0–100 scale.
// Missing history renders an em dash; unfilled width is left-padded with spaces.
func SparklinePercent(samples []float64, width int) string {
	if width <= 0 {
		return ""
	}
	if len(samples) == 0 {
		return "—"
	}
	if len(samples) > width {
		samples = samples[len(samples)-width:]
	}

	var result strings.Builder
	result.Grow(width * 3)
	for range width - len(samples) {
		result.WriteByte(' ')
	}
	for _, sample := range samples {
		if math.IsNaN(sample) || math.IsInf(sample, 0) {
			result.WriteByte(' ')
			continue
		}
		sample = max(0, min(100, sample))
		level := int(math.Round(sample / 100 * float64(len(sparklineLevels)-1)))
		result.WriteRune(sparklineLevels[level])
	}
	return result.String()
}
