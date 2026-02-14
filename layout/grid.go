package layout

import (
	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
)

var _ uikit.Layout = (*Grid)(nil)

// Grid places children in a fixed column grid. If height > 0 it becomes scrollable and clips via SubImage.
type Grid struct {
	uikit.Base
	uikit.Scroller

	children []uikit.Widget

	columns  int
	padX     int
	padY     int
	gapX     int
	gapY     int
	height   int
	contentH int
}

func NewGrid(theme *uikit.Theme) *Grid {
	cfg := uikit.NewWidgetBaseConfig(theme)
	l := &Grid{
		Base:     uikit.NewBase(cfg),
		columns:  2,
		gapX:     theme.SpaceS,
		gapY:     theme.SpaceS,
		Scroller: uikit.NewScroller(),
	}
	l.Base.HeightCalculator = l.heightCalculator

	return l
}

func (l *Grid) heightCalculator() int {
	if l.height == 0 {
		return l.contentH
	}
	return l.height
}

func (l *Grid) Focusable() bool { return false }

func (l *Grid) SetHeight(h int) {
	l.height = h
}

func (l *Grid) SetPadding(x, y int) {
	l.padX = x
	l.padY = y
}

func (l *Grid) SetGap(x, y int) {
	l.gapX = x
	l.gapY = y
}

func (l *Grid) SetColumns(c int) {
	l.columns = c
}

func (l *Grid) Children() []uikit.Widget {
	return l.children
}

func (l *Grid) SetChildren(ws []uikit.Widget) {
	l.children = ws
}

func (l *Grid) Add(ws ...uikit.Widget) {
	l.children = append(l.children, ws...)
}

func (l *Grid) Clear() {
	l.children = nil
}

func (l *Grid) Update(ctx *uikit.Context) {
	if l.height > 0 {
		l.Scroller.Update(ctx, l.Measure(false), l.contentH)
	}
	l.doLayout(ctx)

	for _, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}
		ch.Update(ctx)
	}
}

func (l *Grid) doLayout(ctx *uikit.Context) {
	vp := l.Measure(false)
	cols := l.columns
	if cols <= 0 {
		cols = 2
	}

	innerW := vp.Dx() - l.padX*2
	if innerW < 0 {
		innerW = 0
	}
	cellW := innerW
	if cols > 0 {
		cellW = (innerW - (cols-1)*l.gapX) / cols
		if cellW < 0 {
			cellW = 0
		}
	}

	x := vp.Min.X + l.padX
	y := vp.Min.Y + l.padY - l.Scroller.ScrollY

	l.contentH = l.padY * 2
	rowMaxH := 0
	col := 0
	lastRowCompleted := false
	anyDrawn := false
	for _, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}

		ch.SetFrame(x+(cellW+l.gapX)*col, y, cellW)
		r := ch.Measure(true)
		rowMaxH = max(rowMaxH, r.Dy())

		col += 1
		if col >= cols {
			lastRowCompleted = true
			l.contentH += rowMaxH + l.gapY
			y += rowMaxH + l.gapY
			rowMaxH = 0
			col = 0
		}
	}

	if anyDrawn {
		if !lastRowCompleted {
			l.contentH += rowMaxH
		} else {
			l.contentH -= l.gapY
		}
	}
}

func (l *Grid) Draw(ctx *uikit.Context, dst *ebiten.Image) {
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

func (l *Grid) DrawOverlay(ctx *uikit.Context, dst *ebiten.Image) {
	if !l.IsVisible() {
		return
	}

	for _, ch := range l.children {
		if ow, ok := any(ch).(uikit.OverlayWidget); ok && ow.OverlayActive() {
			ow.DrawOverlay(ctx, dst)
		}
		if ll, ok := any(ch).(interface {
			DrawOverlay(*uikit.Context, *ebiten.Image)
		}); ok {
			ll.DrawOverlay(ctx, dst)
		}
	}
}
