package uikit

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Context holds shared state for all widgets.
type Context struct {
	root    Layout
	theme   *Theme
	ime     IMEBridge
	widgets []Widget
	visible []bool            // widget visible in hierarchy
	clips   []image.Rectangle // widget rects clipped in hierarchy
	focus   int               // -1 means none

	ptr            *PointerStatus
	clickStartRect image.Rectangle
	hasTouch       bool
	prevTouches    map[ebiten.TouchID]struct{}
	dstBounds      image.Rectangle

	// PointerMapper receives a coordinate pair consistent with
	// Game.Layout and returns the corresponding offscreen position.
	//
	// This allows arbitrary pointer projections when the UI has to
	// be drawn to an offscreen first and then projected.
	PointerMapper func(image.Point) image.Point
}

func NewContext(theme *Theme, root Layout, ime IMEBridge) *Context {
	return &Context{
		theme:       theme,
		ime:         ime,
		focus:       -1,
		prevTouches: map[ebiten.TouchID]struct{}{},
		root:        root,
		ptr:         &PointerStatus{},
	}
}

func (c *Context) Theme() *Theme {
	return c.theme
}

// Root returns the root widget (typically a Layout).
func (c *Context) Root() Layout {
	return c.root
}

// SetIMEBridge sets/updates the IME bridge at runtime.
// It also synchronizes the IME visibility with the currently focused widget.
func (c *Context) SetIMEBridge(b IMEBridge) {
	c.ime = b
	c.updateIME(nil, c.Focused())
}

func (c *Context) Add(w Widget) {
	c.root.Add(w)
}

func (c *Context) Focused() Widget {
	if c.focus < 0 || c.focus >= len(c.widgets) {
		return nil
	}

	return c.widgets[c.focus]
}

// Pointer returns the current pointer state in logical pixels.
// On desktop this is the mouse; on mobile this is the active touch.

func (c *Context) rebuildWidgets() {
	c.widgets = c.widgets[:0]
	c.visible = c.visible[:0]
	c.clips = c.clips[:0]

	var walk func(w Widget, clipRect image.Rectangle, hierarchyVisible bool)
	walk = func(w Widget, clipRect image.Rectangle, hierarchyVisible bool) {
		if w == nil {
			return
		}

		clippedRect := w.Measure(false).Intersect(clipRect)

		hierarchyVisible = hierarchyVisible && w.IsVisible()
		c.widgets = append(c.widgets, w)
		c.visible = append(c.visible, hierarchyVisible)
		c.clips = append(c.clips, clippedRect)

		if hw, ok := any(w).(interface{ Children() []Widget }); ok {
			for _, ch := range hw.Children() {
				walk(ch, clippedRect, hierarchyVisible)
			}
		}
	}

	for _, w := range c.root.Children() {
		walk(w, w.Measure(false), c.root.IsVisible())
	}
}

func (c *Context) Pointer() PointerStatus {
	return *c.ptr
}

func (c *Context) SetFocus(w Widget) {
	old := c.Focused()

	// Resolve new focus index (or -1).
	newIdx := -1
	if w != nil {
		for i, ww := range c.widgets {
			if ww == w {
				newIdx = i
				break
			}
		}
	}

	// Emit focus events if changed
	if old != nil && (newIdx != c.focus) {
		old.Dispatch(Event{Widget: old, Type: EventFocusLost})
	}

	c.focus = newIdx
	newW := c.Focused()
	if newW != nil && newW != old {
		newW.Dispatch(Event{Widget: newW, Type: EventFocusGained})
	}

	// IME show/hide based on focused widget.
	c.updateIME(old, newW)
}

func (c *Context) updateIME(oldW, newW Widget) {
	if c.ime == nil {
		return
	}

	// note: IME shown status can't be reliable monitored. Users can call
	// Show()/Hide() on their own, and Android also doesn't expose this
	var opts IMEOptions
	var hadIME, hasIME bool
	if oldW != nil && oldW.IsEnabled() {
		if _, ok := oldW.(IME); ok {
			hadIME = true
		}
	}

	if newW != nil && newW.IsEnabled() {
		if w, ok := newW.(IME); ok {
			hasIME = true
			opts = w.IME()
		}
	}

	if hadIME && !hasIME {
		c.ime.Hide()
	}

	if hasIME {
		inType, imeOpts := opts.AndroidParameters()
		c.ime.Show(inType, imeOpts)
	}
}

func (c *Context) focusNext() {
	if len(c.widgets) == 0 {
		c.SetFocus(nil)
		return
	}
	start := c.focus
	for i := 0; i < len(c.widgets); i++ {
		idx := (start + 1 + i) % len(c.widgets)
		if c.visible[idx] && c.widgets[idx].IsVisible() && c.widgets[idx].IsEnabled() && c.widgets[idx].Focusable() {
			c.SetFocus(c.widgets[idx])
			return
		}
	}
}

func (c *Context) focusPrev() {
	if len(c.widgets) == 0 {
		c.SetFocus(nil)
		return
	}
	start := c.focus
	for i := 0; i < len(c.widgets); i++ {
		idx := start - 1 - i
		for idx < 0 {
			idx += len(c.widgets)
		}
		if c.visible[idx] && c.widgets[idx].IsVisible() && c.widgets[idx].IsEnabled() && c.widgets[idx].Focusable() {
			c.SetFocus(c.widgets[idx])
			return
		}
	}
}

func (c *Context) readPointerSnapshot() {
	c.ptr.IsJustDown = false
	c.ptr.IsJustUp = false
	c.ptr.IsTouch = false

	// Touch tracking (prefer this on mobile; CursorPosition is always (0,0) there).
	touches := ebiten.TouchIDs()
	curr := map[ebiten.TouchID]struct{}{}
	for _, id := range touches {
		curr[id] = struct{}{}
	}

	// Determine transitions
	var justPressed []ebiten.TouchID
	var justReleased []ebiten.TouchID
	for id := range curr {
		if _, ok := c.prevTouches[id]; !ok {
			justPressed = append(justPressed, id)
		}
	}
	for id := range c.prevTouches {
		if _, ok := curr[id]; !ok {
			justReleased = append(justReleased, id)
		}
	}
	c.prevTouches = curr

	// Acquire an active touch when pressed
	if !c.hasTouch && len(justPressed) > 0 {
		c.ptr.TouchID = justPressed[0]
		c.hasTouch = true
		c.ptr.IsJustDown = true
	}

	if c.hasTouch {
		if _, ok := curr[c.ptr.TouchID]; ok {
			c.ptr.IsDown = true
			c.ptr.IsTouch = true
			c.ptr.Position.X, c.ptr.Position.Y = ebiten.TouchPosition(c.ptr.TouchID)
			if c.PointerMapper != nil {
				c.ptr.Position = c.PointerMapper(c.ptr.Position)
			}
		} else {
			c.ptr.IsDown = false
			c.ptr.IsTouch = true
			c.ptr.IsJustUp = true
			c.hasTouch = false
		}

		return
	}

	c.ptr.Position.X, c.ptr.Position.Y = ebiten.CursorPosition()
	if c.PointerMapper != nil {
		c.ptr.Position = c.PointerMapper(c.ptr.Position)
	}
	c.ptr.IsDown = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	c.ptr.IsJustDown = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	c.ptr.IsJustUp = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
}

func (c *Context) widgetHit(w Widget, pos image.Point) bool {
	if h, ok := any(w).(Hittable); ok {
		return h.HitTest(c, pos)
	}

	return pos.In(w.Measure(false))
}

func (c *Context) topmostAt(pos image.Point) Widget {
	for i := len(c.widgets) - 1; i >= 0; i-- {
		w := c.widgets[i]
		if !w.IsVisible() || !w.IsEnabled() || !c.visible[i] {
			continue // ignore if widget not visible
		}
		if !pos.In(c.clips[i]) {
			continue // ignore if pos outside clipped rect (e.g. due to scroll)
		}

		if c.widgetHit(w, pos) {
			return w
		}
	}

	return nil
}

func (c *Context) Update() {
	c.readPointerSnapshot()
	c.root.Update(c)

	c.rebuildWidgets()
	c.root.SetHeight(c.dstBounds.Dy())
	c.root.SetFrame(c.dstBounds.Min.X, c.dstBounds.Min.Y, c.dstBounds.Dx())

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			c.focusPrev()
		} else {
			c.focusNext()
		}
	}

	var target Widget
	if c.ptr.IsJustDown {
		target = c.topmostAt(c.ptr.Position)
		if target != nil && target.Focusable() && target.IsEnabled() {
			c.SetFocus(target)
		} else {
			c.SetFocus(nil)
		}
	}

	var hoverTarget Widget
	if !c.ptr.IsTouch {
		hoverTarget = c.topmostAt(c.ptr.Position)
	}

	for i, w := range c.widgets {
		if !w.IsVisible() || !c.visible[i] {
			continue
		}

		w.SetHovered(hoverTarget == w)

		// Pointer down routed to the chosen target.
		if c.ptr.IsJustDown && target == w && w.IsEnabled() {
			c.clickStartRect = w.Measure(false)
			w.SetPressed(true)
			w.Dispatch(Event{Widget: w, Type: EventPointerDown, Pointer: c.ptr})
		}

		// Pointer up: release + click if pointer ends inside widget.
		if c.ptr.IsJustUp {
			wasPressed := w.IsPressed()
			if wasPressed {
				w.Dispatch(Event{Widget: w, Type: EventPointerUp, Pointer: c.ptr})

				if w.IsEnabled() && c.widgetHit(w, c.ptr.Position) && c.ptr.Position.In(c.clickStartRect) {
					w.Dispatch(Event{Widget: w, Type: EventClick, Pointer: c.ptr})
				}
			}

			w.SetPressed(false)
		}

		w.SetFocused((c.Focused() == w) && w.IsEnabled() && w.Focusable())
	}
}

func (c *Context) Draw(dst *ebiten.Image) {
	if c.root == nil {
		return
	}

	c.dstBounds = dst.Bounds()
	c.root.SetHeight(c.dstBounds.Dy())
	c.root.Draw(c, dst)
	c.root.DrawOverlay(c, dst)
}
