package uikit

import (
	"image"
	"image/color"

	"github.com/erparts/go-uikit/common"
	"github.com/hajimehoshi/ebiten/v2"
)

type WidgetBaseConfig struct {
	Theme       *Theme
	DrawSurface bool
	DrawBorder  bool
	DrawFocus   bool
	DrawInvalid bool
}

func NewWidgetBaseConfig(theme *Theme) *WidgetBaseConfig {
	return &WidgetBaseConfig{
		Theme:       theme,
		DrawSurface: true,
		DrawBorder:  true,
		DrawFocus:   true,
		DrawInvalid: true,
	}
}

// Base contains shared widget state.
type Base struct {
	EventDispatcher

	cfg   *WidgetBaseConfig
	theme *Theme

	HeightCalculator func() int

	rect image.Rectangle

	hovered   bool
	pressed   bool
	focused   bool
	visible   bool
	enabled   bool
	invalid   bool
	errorText string
}

func NewBase(cfg *WidgetBaseConfig) Base {
	return Base{
		EventDispatcher: NewEventDispatcher(),
		cfg:             cfg,
		theme:           cfg.Theme,
		visible:         true,
		enabled:         true,
	}
}

func (b *Base) controlHeight(extended bool) int {
	if b.theme == nil {
		return 0
	}

	h := b.theme.ControlH
	if b.HeightCalculator != nil {
		h = b.HeightCalculator()
	}

	if ok, _ := b.IsInvalid(); ok && extended {
		m := b.theme.ErrorText().Measure(" ")
		h += b.theme.ErrorGap + m.IntHeight()
	}

	return h
}

func (b *Base) Measure(extended bool) image.Rectangle {
	r := common.ChangeRectangleHeight(b.rect, b.controlHeight(extended))
	return r
}

func (b *Base) ErrorRect() image.Rectangle {
	r := b.Measure(false)
	if ok, _ := b.IsInvalid(); !ok {
		return image.Rectangle{}
	}

	m := b.theme.ErrorText().Measure(" ")
	h := b.theme.ErrorGap + m.IntHeight()

	return image.Rect(r.Min.X, r.Max.Y+b.theme.ErrorGap, r.Max.X, r.Max.Y+h)
}

func (b *Base) IsInvalid() (bool, string) {
	return b.invalid, b.errorText
}

func (b *Base) SetInvalid(err string) {
	b.invalid = true
	b.errorText = err
}

func (b *Base) ClearInvalid() {
	b.invalid = false
	b.errorText = ""
}

// SetFrame sets the widget position (x,y) and width (w).
// The final height is computed from the theme and the widget state (e.g. error line, variable-height controls).
func (b *Base) SetFrame(x, y, w int) {
	if w < 0 {
		w = 0
	}

	b.rect = image.Rect(x, y, x+w, y+b.requiredHeight())
}

func (b *Base) requiredHeight() int {
	h := b.controlHeight(true)
	if ok, _ := b.IsInvalid(); ok {
		m := b.theme.ErrorText().Measure(" ")
		h += b.theme.ErrorGap + m.IntHeight()
	}

	return h
}

func (b *Base) IsHovered() bool   { return b.hovered }
func (b *Base) SetHovered(v bool) { b.hovered = v }

func (b *Base) IsPressed() bool   { return b.pressed }
func (b *Base) SetPressed(v bool) { b.pressed = v }

func (b *Base) IsFocused() bool   { return b.focused }
func (b *Base) SetFocused(v bool) { b.focused = v }

func (b *Base) IsEnabled() bool   { return b.enabled }
func (b *Base) SetEnabled(v bool) { b.enabled = v }

func (b *Base) IsVisible() bool   { return b.visible }
func (b *Base) SetVisible(v bool) { b.visible = v }

func (c *Base) Draw(ctx *Context, dst *ebiten.Image) image.Rectangle {
	r := c.Measure(false)
	c.DrawSurface(ctx, dst, r)
	c.DrawBorder(ctx, dst, r)
	c.DrawFocus(ctx, dst, r)
	c.DrawInvalid(ctx, dst, r)
	return r
}

func (c *Base) DrawSurface(ctx *Context, dst *ebiten.Image, r image.Rectangle) {
	if !c.cfg.DrawSurface {
		return
	}

	bg := ctx.Theme().SurfaceColor
	if !c.enabled {
		bg = ctx.Theme().SurfacePressedColor
	} else if c.pressed {
		bg = ctx.Theme().SurfacePressedColor
	} else if c.hovered {
		bg = ctx.Theme().SurfaceHoverColor
	}

	r = r.Sub(dst.Bounds().Min)
	drawRoundedRect(dst, r, ctx.Theme().Radius, bg)
}

func (c *Base) DrawBorder(ctx *Context, dst *ebiten.Image, r image.Rectangle) {
	if !c.cfg.DrawBorder {
		return
	}

	border := ctx.Theme().BorderColor
	if !c.enabled {
		border = ctx.Theme().DisabledColor
	}
	if c.invalid {
		border = ctx.Theme().ErrorBorderColor
	}

	r = r.Sub(dst.Bounds().Min)
	drawRoundedBorder(dst, r, ctx.Theme().Radius, ctx.Theme().BorderW, border)
}

func (c *Base) DrawFocus(ctx *Context, dst *ebiten.Image, r image.Rectangle) {
	if !c.cfg.DrawFocus {
		return
	}

	if !c.focused || !c.enabled {
		return
	}

	r = r.Sub(dst.Bounds().Min)
	drawRoundedBorder(dst, r, ctx.Theme().Radius, ctx.Theme().FocusRingW, ctx.Theme().FocusColor)
}

func (c *Base) DrawInvalid(ctx *Context, dst *ebiten.Image, r image.Rectangle) {
	if !c.cfg.DrawInvalid {
		return
	}

	if !c.invalid {
		return
	}

	err := c.ErrorRect()
	drawErrorText(ctx, dst, err, c.errorText)
}

func (c *Base) Theme() *Theme {
	return c.theme
}

func (c *Base) DrawRoundedRect(dst *ebiten.Image, r image.Rectangle, radius int, col color.RGBA) {
	drawRoundedRect(dst, r, radius, col)
}

func (c *Base) DrawRoundedBorder(dst *ebiten.Image, r image.Rectangle, radius int, borderW int, col color.RGBA) {
	drawRoundedBorder(dst, r, radius, borderW, col)
}
