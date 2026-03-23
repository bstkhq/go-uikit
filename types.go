package uikit

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Widget interface {
	SetFrame(x, y, w int)

	IsHovered() bool
	SetHovered(bool)
	IsPressed() bool
	SetPressed(bool)
	IsFocused() bool
	SetFocused(bool)
	IsEnabled() bool
	SetEnabled(bool)
	IsVisible() bool
	SetVisible(bool)

	Focusable() bool

	Measure(bool) image.Rectangle
	Update(ctx *Context)
	Draw(ctx *Context, dst *ebiten.Image)

	On(t EventType, cb EventHandler, clear bool)
	Dispatch(e Event)
}

// OverlayWidget can draw an overlay above all other widgets (e.g. Select dropdown).
type OverlayWidget interface {
	OverlayActive() bool
	DrawOverlay(ctx *Context, dst *ebiten.Image)
}

type ValidableWidget interface {
	IsValidable() bool
	IsInvalid() (bool, string)
	SetInvalid(err string)
	ClearInvalid()
}

// Hittable allows a widget to extend its clickable area beyond Base.ControlRect.
type Hittable interface {
	HitTest(ctx *Context, pos image.Point) bool
}

// Layout is a Widget that owns children.
type Layout interface {
	Widget
	DrawOverlay(ctx *Context, dst *ebiten.Image)
	SetHeight(int)
	SetPadding(int, int)
	Children() []Widget
	SetChildren([]Widget)
	Add(...Widget)
	Clear()
}

type PointerStatus struct {
	Position   image.Point
	IsDown     bool
	IsJustDown bool
	IsJustUp   bool
	IsTouch    bool
	TouchID    ebiten.TouchID
}
