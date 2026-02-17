package layout

import (
	"cmp"
	"math/rand/v2"
	"testing"
)

func TestRowStackPattern(t *testing.T) {
	var buff []int
	for i := range 500 {
		var weights []float64
		for range 1 + rand.IntN(16) {
			weights = append(weights, rand.Float64())
		}
		rsp := rowStackPattern{Weights: weights}
		rsp.normalize()
		gap := rand.IntN(56)
		wsum := sum(rsp.Weights)
		if wsum > 1.0000001 || wsum < 0.9999999 {
			t.Fatalf("expected sum = 1.0, got %f", wsum)
		}
		minWidth := gap * (len(weights) - 1)
		totalWidth := int(float64(minWidth) * (0.99 + rand.Float64()*6))
		totalWidth = max(minWidth, totalWidth)
		buff = rsp.computeWidths(totalWidth, gap, buff)
		actualTotal := sum(buff) + gap*(len(weights)-1)
		if actualTotal != totalWidth {
			t.Fatalf("test#%d: expected %d, got %d | gap: %d, weights %v", i, totalWidth, actualTotal, gap, rsp.Weights)
		}
	}
}

func sum[T cmp.Ordered](vs []T) T {
	var s T
	for _, v := range vs {
		s += v
	}
	return s
}
