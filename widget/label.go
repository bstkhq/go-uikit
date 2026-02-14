package widget

import (
	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

var _ uikit.Widget = (*Label)(nil)

type Label struct {
	uikit.Base
	text      string
	textFunc  func() string
	modifiers []TextModifier

	lastHeight int
	refWidth   int
}

func NewLabel(theme *uikit.Theme, text string) *Label {
	cfg := uikit.NewWidgetBaseConfig(theme)
	cfg.DrawSurface = false
	cfg.DrawBorder = false

	w := &Label{
		Base:     uikit.NewBase(cfg),
		text:     text,
		refWidth: -1,
	}
	w.Base.HeightCalculator = w.heightCalculator

	return w
}

func (w *Label) heightCalculator() int {
	return w.lastHeight
}

func (w *Label) SetFrame(x, y, width int) {
	if w.refWidth != width {
		w.refreshHeight(width)
	}
	w.Base.SetFrame(x, y, width)
}

func (w *Label) refreshHeight(width int) {
	w.refWidth = width
	renderer := w.textRenderer(w.Base.Theme())
	w.lastHeight = renderer.MeasureWithWrap(w.text, w.refWidth).IntHeight()
}

func (w *Label) Focusable() bool {
	return false
}

func (w *Label) SetText(s string) {
	if w.text != s {
		w.text = s
		w.refWidth = -1
	}
}

func (w *Label) SetTextFunc(fn func() string) {
	w.textFunc = fn
}

func (w *Label) SetTextModifiers(mods ...TextModifier) {
	w.modifiers = mods
	w.refWidth = -1
}

func (w *Label) Update(ctx *uikit.Context) {
	if w.textFunc != nil {
		text := w.textFunc()
		if text != w.text {
			w.text = text
			w.refWidth = -1
		}
	}

	if w.refWidth < 0 {
		w.refreshHeight(w.Measure(false).Dx())
	}
}

func (w *Label) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	r := w.Base.Draw(ctx, dst)

	renderer := w.textRenderer(ctx.Theme())
	x := renderer.GetAlign().Horz().GetHorzAnchor(r.Min.X, r.Max.X)

	renderer.DrawWithWrap(dst, w.text, x, r.Min.Y+(r.Dy()/2), r.Dx())
}

func (w *Label) textRenderer(theme *uikit.Theme) *etxt.Renderer {
	renderer := theme.Text()
	renderer.SetColor(theme.TextColor)
	renderer.SetAlign(etxt.Left | etxt.VertCenter)
	for _, mod := range w.modifiers {
		mod(theme, renderer)
	}
	return renderer
}
