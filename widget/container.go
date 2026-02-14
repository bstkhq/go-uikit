package widget

import (
	"image"

	"github.com/erparts/go-uikit"
	"github.com/erparts/go-uikit/common"
	"github.com/hajimehoshi/ebiten/v2"
)

var _ uikit.Widget = (*Container)(nil)

// Container is an empty widget that lets you render custom content inside a themed box.
// It still participates in focus/invalid layout like any other widget.
type Container struct {
	uikit.Base
	height int

	OnUpdate func(ctx *uikit.Context, content image.Rectangle)
	OnDraw   func(ctx *uikit.Context, dst *ebiten.Image)
}

func NewContainer(theme *uikit.Theme) *Container {
	cfg := uikit.NewWidgetBaseConfig(theme)

	w := &Container{}
	w.Base = uikit.NewBase(cfg)
	w.Base.HeightCalculator = func() int {
		return w.height
	}

	return w
}

func (w *Container) SetHeight(h int) {
	w.height = h
}

func (w *Container) Focusable() bool {
	return false
}

func (w *Container) Update(ctx *uikit.Context) {
	if w.OnUpdate != nil {
		theme := ctx.Theme()
		w.OnUpdate(ctx, common.Inset(w.Measure(false), theme.PadX, theme.PadY))
	}
}

func (w *Container) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	r := w.Base.Draw(ctx, dst)

	content := common.Inset(r, ctx.Theme().PadX, ctx.Theme().PadY)
	if w.OnDraw != nil {
		w.OnDraw(ctx, dst.SubImage(content).(*ebiten.Image))
	}
}
