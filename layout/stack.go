package layout

import (
	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
)

var _ uikit.Layout = (*Stack)(nil)

// Stack places children vertically. If height > 0 it becomes scrollable and clips via SubImage.
type Stack struct {
	uikit.Base
	uikit.Scroller

	children []uikit.Widget

	padX int
	padY int
	gap  int

	height   int
	contentH int
}

func NewStack(theme *uikit.Theme) *Stack {
	cfg := uikit.NewWidgetBaseConfig(theme)
	l := &Stack{
		Base:     uikit.NewBase(cfg),
		Scroller: uikit.NewScroller(),
		gap:      theme.SpaceS,
	}
	l.Base.HeightCalculator = l.heightCalculator

	return l
}

func (l *Stack) heightCalculator() int {
	if l.height == 0 {
		return l.contentH
	}
	return l.height
}

func (l *Stack) Focusable() bool { return false }

// SetHeight sets the viewport height. Use 0 for unlimited (no scroll, no clipping).
func (l *Stack) SetHeight(h int) {
	l.height = h
}

func (l *Stack) SetPadding(x, y int) {
	l.padX = x
	l.padY = y
}

func (l *Stack) SetGap(v int) {
	l.gap = v
}

func (l *Stack) Children() []uikit.Widget {
	return l.children
}
func (l *Stack) SetChildren(ws []uikit.Widget) {
	l.children = ws
}
func (l *Stack) Add(ws ...uikit.Widget) {
	l.children = append(l.children, ws...)
}

func (l *Stack) Clear() {
	l.children = nil
}

func (l *Stack) Update(ctx *uikit.Context) {
	r := l.Measure(false)
	if l.height > 0 {
		l.Scroller.Update(ctx, r, l.contentH)
	}
	l.doLayout(ctx)

	for _, w := range l.children {
		if !w.IsVisible() {
			continue
		}

		w.Update(ctx)
	}
}

func (l *Stack) doLayout(ctx *uikit.Context) {
	vp := l.Measure(false)
	x := vp.Min.X + l.padX
	y := vp.Min.Y + l.padY - l.Scroller.ScrollY
	w := max(vp.Dx()-l.padX*2, 0)

	l.contentH = l.padY * 2
	var anyDrawn bool
	for _, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}

		ch.SetFrame(x, y, w)
		advance := ch.Measure(true).Dy() + l.gap
		y += advance
		l.contentH += advance
		anyDrawn = true
	}

	if anyDrawn {
		l.contentH -= l.gap
	}
}

func (l *Stack) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	if !l.IsVisible() {
		return
	}

	r := l.Measure(false)
	sub := dst.SubImage(r).(*ebiten.Image)
	for _, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}
		ch.Draw(ctx, sub)
	}
	if l.height > 0 {
		l.Scroller.DrawBar(sub, ctx.Theme(), sub.Bounds().Dx(), sub.Bounds().Dy(), l.contentH)
	}
}

func (l *Stack) DrawOverlay(ctx *uikit.Context, dst *ebiten.Image) {
	if !l.IsVisible() {
		return
	}

	// Overlay should escape clipping -> draw on dst (not on subimage)
	for _, ch := range l.children {
		if ow, ok := any(ch).(uikit.OverlayWidget); ok && ow.OverlayActive() {
			ow.DrawOverlay(ctx, dst)
		}
		// Nested layouts will propagate overlays naturally.
		if ll, ok := any(ch).(interface {
			DrawOverlay(*uikit.Context, *ebiten.Image)
		}); ok {
			ll.DrawOverlay(ctx, dst)
		}
	}
}
