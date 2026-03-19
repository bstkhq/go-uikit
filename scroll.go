package uikit

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ScrollbarMode controls when the scrollbar is rendered.
type ScrollbarMode int

const (
	ScrollbarNever ScrollbarMode = iota
	ScrollbarAlways
	// ScrollbarOnMove shows the bar only while scrolling/dragging, and for a short
	// time after the last scroll input.
	ScrollbarOnMove
)

// Scroller is a small helper that manages vertical scrolling and a simple scrollbar.
// It is intentionally simple and relies on clipping via SubImage when drawing.
type Scroller struct {
	ScrollY int

	// Drag state
	dragging bool
	lastPY   int

	// For ScrollbarOnMove
	showTicks int

	Scrollbar ScrollbarMode
}

func NewScroller() Scroller {
	return Scroller{
		Scrollbar: ScrollbarOnMove,
	}
}

func (s *Scroller) IsScrolling() bool { return s.dragging || s.showTicks > 0 }

// Update updates scrolling using wheel + drag/touch, only if the pointer is inside viewport.
// contentH is the full scrollable content height in pixels.
func (s *Scroller) Update(ctx *Context, viewport image.Rectangle, contentH int) {
	if viewport.Dx() <= 0 || viewport.Dy() <= 0 {
		return
	}

	ptr := ctx.Pointer()
	inside := ptr.Position.In(viewport)

	changed := false

	// Wheel (desktop)
	_, wy := ebiten.Wheel()
	if wy != 0 && inside {
		step := int(math.Round(float64(ctx.Theme().ControlH) * 0.65))
		if step < 10 {
			step = 10
		}
		s.ScrollY -= int(math.Round(wy * float64(step)))
		changed = true
	}

	// Drag (mouse/touch)
	if ptr.IsJustDown && inside {
		s.dragging = true
		s.lastPY = ptr.Position.Y
		s.showTicks = 18
	}

	if s.dragging && ptr.IsDown {
		dy := ptr.Position.Y - s.lastPY
		s.ScrollY -= dy
		s.lastPY = ptr.Position.Y
		if dy != 0 {
			changed = true
		}
	}

	if ptr.IsJustUp {
		s.dragging = false
	}

	s.Clamp(viewport.Dy(), contentH)

	if s.Scrollbar == ScrollbarOnMove {
		if changed {
			s.showTicks = 18
		} else if s.showTicks > 0 {
			s.showTicks--
		}
	}
}

// Clamp clamps ScrollY to the valid range for the given viewport height and content height.
func (s *Scroller) Clamp(viewportH int, contentH int) {
	max := contentH - viewportH
	if max < 0 {
		max = 0
	}
	if s.ScrollY < 0 {
		s.ScrollY = 0
	}
	if s.ScrollY > max {
		s.ScrollY = max
	}
}

// DrawBar draws a simple vertical scrollbar inside a clipped target (dst should already be a SubImage of the viewport).
// viewportW/H should match dst's size.
func (s *Scroller) DrawBar(dst *ebiten.Image, theme *Theme, viewportW, viewportH, contentH int) {
	if viewportW <= 0 || viewportH <= 0 {
		return
	}

	if contentH <= viewportH {
		return
	}

	show := false
	switch s.Scrollbar {
	case ScrollbarNever:
		show = false
	case ScrollbarAlways:
		show = true
	case ScrollbarOnMove:
		show = s.IsScrolling()
	}

	if !show {
		return
	}

	trackW := int(math.Max(3, float64(theme.BorderW)))
	trackX := viewportW - trackW
	trackH := viewportH

	thumbH := int(math.Max(12, float64(trackH)*float64(trackH)/float64(contentH)))
	maxScroll := contentH - trackH

	thumbY := 0
	if maxScroll > 0 {
		thumbY = int(math.Round(float64(trackH-thumbH) * float64(s.ScrollY) / float64(maxScroll)))
	}

	ox := dst.Bounds().Min.X
	oy := dst.Bounds().Min.Y

	vector.DrawFilledRect(dst, float32(trackX+ox), float32(oy), float32(trackW), float32(trackH), theme.BorderColor, false)
	vector.DrawFilledRect(dst, float32(trackX+ox), float32(oy+thumbY), float32(trackW), float32(thumbH), theme.FocusColor, false)
}
