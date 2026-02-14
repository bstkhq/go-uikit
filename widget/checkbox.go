package widget

import (
	"image"
	"math"

	"github.com/erparts/go-uikit"
	"github.com/erparts/go-uikit/common"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
)

var _ uikit.Widget = (*Checkbox)(nil)

type Checkbox struct {
	uikit.Base

	label   string
	checked bool

	lastHeight int
	refWidth   int
}

func NewCheckbox(theme *uikit.Theme, label string) *Checkbox {
	cfg := uikit.NewWidgetBaseConfig(theme)

	w := &Checkbox{
		Base:  uikit.NewBase(cfg),
		label: label,
	}
	w.Base.HeightCalculator = w.heightCalculator

	// Clicking anywhere on the widget triggers toggle (Base must emit EventClick).
	w.Base.On(uikit.EventClick, w.onClick, false)

	return w
}

func (w *Checkbox) heightCalculator() int {
	return w.lastHeight
}

func (w *Checkbox) SetFrame(x, y, width int) {
	if width != w.refWidth {
		w.refreshHeight(width)
	}
	w.Base.SetFrame(x, y, width)
}

func (w *Checkbox) refreshHeight(width int) {
	w.refWidth = width
	theme := w.Base.Theme()
	boxSize, boxHorzIntsp, padX, padY := w.boxMetrics(theme)
	maxLineLen := max(w.refWidth-(boxSize+boxHorzIntsp+padX*2), 0)
	w.lastHeight = theme.Text().MeasureWithWrap(w.label, maxLineLen).IntHeight()
	w.lastHeight = max(w.lastHeight+padY*2, theme.ControlH)
}

func (w *Checkbox) Focusable() bool { return true }

func (w *Checkbox) SetChecked(v bool) {
	if w.checked == v {
		return
	}
	w.checked = v
	w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventValueChange})
}

func (w *Checkbox) Checked() bool { return w.checked }

func (w *Checkbox) onClick(e uikit.Event) bool {
	if !w.IsEnabled() {
		return false
	}

	if e.Type == uikit.EventClick {
		w.SetChecked(!w.Checked())
	}

	return false
}

func (w *Checkbox) Update(ctx *uikit.Context) {
	if !w.IsEnabled() {
		return
	}

	// Keyboard toggle
	if w.IsFocused() && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		w.SetChecked(!w.Checked())
	}
}

func (w *Checkbox) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	r := w.Base.Draw(ctx, dst)
	dr := r.Sub(dst.Bounds().Min)

	theme := ctx.Theme()
	boxSize, boxHorzIntsp, padX, padY := w.boxMetrics(ctx.Theme())
	content := common.Inset(dr, padX, padY)
	boxY := dr.Min.Y + (dr.Dy()-boxSize)/2
	box := image.Rect(content.Min.X, boxY, content.Min.X+boxSize, boxY+boxSize)

	// Colors
	bg := theme.BackgroundColor
	border := theme.BorderColor
	checkCol := theme.FocusColor
	textCol := theme.TextColor

	if !w.IsEnabled() {
		border = theme.DisabledColor
		checkCol = theme.DisabledColor
		textCol = theme.DisabledColor
	}

	radius := int(math.Round(float64(boxSize) * 0.22))
	if radius < 2 {
		radius = 2
	}

	if w.IsEnabled() {
		if w.IsPressed() {
			bg = theme.FocusColor
		} else if w.IsHovered() {
			bg = theme.BorderColor
		}
	}

	w.Base.DrawRoundedRect(dst, box, radius, bg)
	w.Base.DrawRoundedBorder(dst, box, radius, theme.BorderW, border)

	if w.checked {
		boxMin := r.Min.Add(image.Pt(padX, (dr.Dy()-boxSize)/2))
		x1 := float32(boxMin.X) + float32(boxSize)*0.22
		y1 := float32(boxMin.Y) + float32(boxSize)*0.55
		x2 := float32(boxMin.X) + float32(boxSize)*0.42
		y2 := float32(boxMin.Y) + float32(boxSize)*0.73
		x3 := float32(boxMin.X) + float32(boxSize)*0.78
		y3 := float32(boxMin.Y) + float32(boxSize)*0.30

		strokeW := float32(theme.BorderW)
		if strokeW < 2 {
			strokeW = 2
		}
		// Slightly thicker for better readability at small sizes
		if strokeW < float32(boxSize)/8 {
			strokeW = float32(boxSize) / 8
		}

		vector.StrokeLine(dst, x1, y1, x2, y2, strokeW, checkCol, true)
		vector.StrokeLine(dst, x2, y2, x3, y3, strokeW, checkCol, true)
	}

	t := theme.Text()
	t.SetColor(textCol)
	t.SetAlign(etxt.Left | etxt.VertCenter)
	leftOffset := boxSize + boxHorzIntsp + padX
	maxLineLen := max(r.Dx()-(leftOffset+padX), 0)
	t.DrawWithWrap(dst, w.label, r.Min.X+leftOffset, r.Min.Y+r.Dy()/2, maxLineLen)
}

func (w *Checkbox) boxMetrics(theme *uikit.Theme) (size, horzInterspace, padX, padY int) {
	boxSize := theme.CheckSize
	if boxSize < 12 {
		boxSize = 12
	}
	return boxSize, theme.SpaceS, theme.PadX, theme.PadY
}
