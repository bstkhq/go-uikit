package widget

import (
	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
)

var _ uikit.Widget = (*Button)(nil)

// Button is a clickable control with hover/pressed/disabled visuals.
// - Click triggers on pointer release inside the widget.
// - Enter/Space triggers click when focused.
type Button struct {
	uikit.Base

	label   string
	OnClick func()

	// internal: tracks if the press started inside this widget
	pressedInside bool
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

// fireClick dispatches a click event and calls OnClick handler.
func (w *Button) fireClick() {
	w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventClick})
	if w.OnClick != nil {
		w.OnClick()
	}
}

func (w *Button) Update(ctx *uikit.Context) {
	if !w.IsEnabled() {
		w.pressedInside = false
		return
	}

	if w.IsFocused() && (inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)) {
		w.fireClick()
		return
	}

	ptr := ctx.Pointer()
	inside := ptr.Position.In(w.Measure(false))

	// Start press inside
	if ptr.IsJustDown && inside {
		w.pressedInside = true
	}

	if ptr.IsJustDown {
		if w.pressedInside && inside {
			w.fireClick()
		}

		w.pressedInside = false
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
