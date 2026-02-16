package layout

import (
	"math"
	"slices"

	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
)

type Align int

func (a Align) Offset(itemSize, containerSize int) int {
	switch a {
	case AlignMiddle:
		return (containerSize - itemSize) / 2
	case AlignEnd:
		return containerSize - itemSize
	default: // assume AlignStart
		return 0
	}
}

const (
	AlignStart  Align = -1
	AlignMiddle Align = 0
	AlignEnd    Align = 1
)

type rowStackPattern struct {
	Weights []float64
}

func (rsp *rowStackPattern) normalize() {
	if len(rsp.Weights) == 0 {
		return
	}

	var sum float64
	for i, weight := range rsp.Weights {
		if weight < 0 {
			rsp.Weights[i] = 0
		} else {
			sum += weight
		}
	}

	if sum > 0 {
		for i, weight := range rsp.Weights {
			rsp.Weights[i] = weight / sum
		}
	} else {
		div := float64(len(rsp.Weights))
		for i := range rsp.Weights {
			rsp.Weights[i] = 1.0 / div
		}
	}
}

func (rsp *rowStackPattern) computeWidths(totalWidth, gap int, widthsBuffer []int) []int {
	widthsBuffer = widthsBuffer[:0]
	if len(rsp.Weights) <= 1 {
		widthsBuffer = append(widthsBuffer, totalWidth)
		return widthsBuffer
	}

	totalWidth -= gap * max(len(rsp.Weights)-1, 0)
	if totalWidth <= 0 {
		for range len(rsp.Weights) {
			widthsBuffer = append(widthsBuffer, 0)
		}
		return widthsBuffer
	}

	tw := float64(totalWidth)
	accWidth := 0
	accWeight := 0.0
	lastNonZero := -1
	for i, weight := range rsp.Weights {
		if weight == 0 {
			widthsBuffer = append(widthsBuffer, 0)
			continue
		}

		accWeight += weight
		expWidth := int(math.Round(accWeight * tw))
		widthInt := expWidth - accWidth
		accWidth += widthInt
		if widthInt > 0 || lastNonZero == -1 {
			lastNonZero = i
		}
		widthsBuffer = append(widthsBuffer, widthInt)
	}

	if accWidth > totalWidth {
		widthsBuffer[lastNonZero] -= 1
	} else if accWidth < totalWidth {
		widthsBuffer[lastNonZero] += 1
	}

	return widthsBuffer
}

type helperFrameAlign struct {
	Widget uikit.Widget
	X, Y   int
	W, H   int
}

var _ uikit.Layout = (*RowStack)(nil)

// RowStack is a top to bottom layout where each row can have a
// different number of widgets and a different amount of space
// for each.
type RowStack struct {
	uikit.Base
	uikit.Scroller

	patterns   map[int]rowStackPattern
	children   []uikit.Widget
	padX, padY int
	height     int
	contentH   int
	hGap, vGap int
	align      Align

	widthsBuffer []int
	framesBuffer []helperFrameAlign

	// BeforeUpdate can be used to adjust children or configuration
	// before layouting.
	BeforeUpdate func(*uikit.Context, *RowStack)
}

// NewRowStack creates a RowStack with the given default pattern.
func NewRowStack(theme *uikit.Theme, widthWeights ...float64) *RowStack {
	defaultPattern := rowStackPattern{Weights: widthWeights}
	defaultPattern.normalize()

	cfg := uikit.NewWidgetBaseConfig(theme)
	l := &RowStack{
		Base:     uikit.NewBase(cfg),
		Scroller: uikit.NewScroller(),
		patterns: map[int]rowStackPattern{-1: defaultPattern},
		hGap:     theme.SpaceS,
		vGap:     theme.SpaceS,
	}

	l.Base.HeightCalculator = l.heightCalculator
	return l
}

func (l *RowStack) heightCalculator() int {
	if l.height == 0 {
		return l.contentH
	}
	return l.height
}

// DefaultPattern returns the weights of the default row pattern.
func (l *RowStack) DefaultPattern() []float64 {
	pattern := l.patterns[-1]
	return pattern.Weights
}

// SetRowPattern sets the widget distribution pattern of a specific row.
// If rowIndex < 0, the given pattern is set as the default.
func (l *RowStack) SetRowPattern(rowIndex int, widthWeights ...float64) {
	pattern := rowStackPattern{Weights: slices.Clone(widthWeights)}
	pattern.normalize()
	rowIndex = max(rowIndex, -1)
	l.patterns[rowIndex] = pattern
}

// SetGap sets the horizontal gap between items and the vertical gap
// between rows.
func (l *RowStack) SetGap(horzGap, vertGap int) {
	l.hGap, l.vGap = horzGap, vertGap
}

// SetItemAlign sets the vertical align for items within a row.
func (l *RowStack) SetAlign(align Align) {
	l.align = align
}

func (l *RowStack) Focusable() bool { return false }

func (l *RowStack) SetHeight(h int) {
	l.height = h
}

func (l *RowStack) SetPadding(x, y int) {
	l.padX = x
	l.padY = y
}

func (l *RowStack) Children() []uikit.Widget {
	return l.children
}

func (l *RowStack) SetChildren(ws []uikit.Widget) {
	l.children = ws
}

func (l *RowStack) Add(ws ...uikit.Widget) {
	l.children = append(l.children, ws...)
}

func (l *RowStack) Clear() {
	l.children = l.children[:0]
}

func (l *RowStack) Update(ctx *uikit.Context) {
	if l.BeforeUpdate != nil {
		l.BeforeUpdate(ctx, l)
	}
	l.doLayout(ctx)
	if l.height > 0 {
		l.Scroller.Update(ctx, l.Measure(false), l.height)
	}

	for _, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}
		ch.Update(ctx)
	}
}

func (l *RowStack) doLayout(ctx *uikit.Context) {
	area := l.Measure(false)

	ox := area.Min.X + l.padX
	y := area.Min.Y + l.padY - l.Scroller.ScrollY
	contentWidth := max(area.Dx()-l.padX*2, 0)

	basePattern := l.patterns[-1]

	l.contentH = l.padY * 2
	var anyVisible bool
	rowIndex := 0
	childIndex := 0
	onBasePattern := false
	for {
		pattern, ok := l.patterns[rowIndex]
		refreshWidths := ok || !onBasePattern
		if !onBasePattern && !ok {
			pattern = basePattern
		}
		onBasePattern = !ok
		if refreshWidths {
			l.widthsBuffer = pattern.computeWidths(contentWidth, l.hGap, l.widthsBuffer)
		}

		maxHeight := 0
		colIndex := 0
		x := ox
		l.framesBuffer = l.framesBuffer[:0]
		children := l.children[childIndex:]
		for _, child := range children {
			childIndex += 1
			if !child.IsVisible() {
				continue
			}

			anyVisible = true
			width := l.widthsBuffer[colIndex]
			child.SetFrame(x, y, width)
			height := child.Measure(true).Dy()
			l.framesBuffer = append(l.framesBuffer, helperFrameAlign{Widget: child, X: x, Y: y, W: width, H: height})
			maxHeight = max(maxHeight, height)
			x += width + l.hGap
			colIndex += 1
			if colIndex >= len(l.widthsBuffer) {
				break
			}
		}
		if colIndex == 0 {
			break // no more children
		}

		// apply vertical centering if required
		if l.align != AlignStart {
			for _, f := range l.framesBuffer {
				offset := l.align.Offset(f.H, maxHeight)
				if offset != 0 {
					f.Widget.SetFrame(f.X, f.Y+offset, f.W)
				}
			}
		}

		l.contentH += maxHeight + l.vGap
		y += maxHeight + l.vGap

		rowIndex += 1
	}

	if anyVisible {
		l.contentH -= l.vGap
	}
}

func (l *RowStack) Draw(ctx *uikit.Context, dst *ebiten.Image) {
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

type OverlayDrawer interface {
	DrawOverlay(*uikit.Context, *ebiten.Image)
}

func (l *RowStack) DrawOverlay(ctx *uikit.Context, dst *ebiten.Image) {
	if !l.IsVisible() {
		return
	}

	for _, child := range l.children {
		switch overlay := child.(type) {
		case uikit.OverlayWidget:
			if overlay.OverlayActive() {
				overlay.DrawOverlay(ctx, dst)
			}
		case OverlayDrawer: // TODO: maybe simplify to only uikit.OverlayWidget and adjust widget.Select?
			overlay.DrawOverlay(ctx, dst)
		}
	}
}
