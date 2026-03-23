package widget

import (
	"github.com/bstkhq/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
)

var _ uikit.Widget = (*Button)(nil)

// Button is a clickable control with hover/pressed/disabled visuals.
//
// In addition to standard [EventClick] triggering, pressing Enter and
// Space will also dispatch a click event for buttons.
type Button struct {
	uikit.Base
	label string
}

func NewButton(theme *uikit.Theme, label string) *Button {
	cfg := uikit.NewWidgetBaseConfig(theme)

	b := &Button{
		Base:  uikit.NewBase(cfg),
		label: label,
	}

	return b
}

func (w *Button) Focusable() bool { return true }

func (w *Button) SetLabel(s string) {
	w.label = s
}

func (w *Button) Update(ctx *uikit.Context) {
	if !w.IsEnabled() {
		return
	}

	if w.IsFocused() && (inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)) {
		w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventClick})
		return
	}
}

func (w *Button) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	r := w.Base.Draw(ctx, dst)
	dr := r.Sub(dst.Bounds().Min)

	theme := ctx.Theme()
	if w.IsEnabled() {
		if w.IsPressed() {
			w.DrawRoundedRect(dst, dr, theme.Radius, theme.FocusColor)
		} else if w.IsHovered() {
			w.DrawRoundedRect(dst, dr, theme.Radius, theme.BorderColor)
		}
	}

	col := theme.TextColor
	if !w.IsEnabled() {
		col = theme.DisabledColor
	}

	t := theme.Text()
	t.SetColor(col)
	t.SetAlign(etxt.Center)

	offY := 0
	if w.IsEnabled() && w.IsPressed() {
		offY = 0
	}

	t.Draw(dst, w.label, r.Min.X+r.Dx()/2, r.Min.Y+r.Dy()/2+offY)
}
