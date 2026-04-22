package common

import (
	"image"
	"math"
)

// PointerMapper is a reference implementation for uikit.Context.PointerMapper
// that allows adjusting pointer coordinates when drawing UI from an offscreen.
type PointerMapper struct {
	Target    image.Rectangle
	Offscreen image.Rectangle
}

// Project matches the function signature for uikit.Context.PointerMapper.
func (m *PointerMapper) Project(ptr image.Point) image.Point {
	scale := func(from, to, at int) int {
		if from == 0 {
			return at
		}
		s := float64(to) / float64(from)
		return int(math.Round(s * float64(at)))
	}

	ptr = ptr.Sub(m.Target.Min)
	ptr.X = scale(m.Target.Dx(), m.Offscreen.Dx(), ptr.X)
	ptr.Y = scale(m.Target.Dy(), m.Offscreen.Dy(), ptr.Y)
	return ptr.Add(m.Offscreen.Min)
}
