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
	focus   int // -1 means none

	ptr         *PointerStatus
	hasTouch    bool
	prevTouches map[ebiten.TouchID]struct{}
	dstBounds   image.Rectangle
}

func NewContext(theme *Theme, root Layout, ime IMEBridge) *Context {
	return &Context{
		theme:       theme,
		ime:         ime,
		focus:       -1,
		prevTouches: map[ebiten.TouchID]struct{}{},
		root:        root,
		widgets:     []Widget{root},
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
	c.updateIMEForce(c.Focused())
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
	var walk func(w Widget)
	walk = func(w Widget) {
		if w == nil {
			return
		}

		c.widgets = append(c.widgets, w)

		if hw, ok := any(w).(interface{ Children() []Widget }); ok {
			for _, ch := range hw.Children() {
				walk(ch)
			}
		}
	}

	for _, w := range c.root.Children() {
		walk(w)
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

	oldWants := false
	if oldW != nil {
		if wi, ok := any(oldW).(WantsIME); ok && wi.WantsIME() {
			oldWants = true
		}
	}
	newWants := false
	if newW != nil {
		if wi, ok := any(newW).(WantsIME); ok && wi.WantsIME() {
			newWants = true
		}
	}

	// Only issue calls on state transitions.
	if oldWants && !newWants {
		c.ime.Hide()
	}
	if !oldWants && newWants {
		c.ime.Show()
	}
}

func (c *Context) updateIMEForce(focused Widget) {
	if c.ime == nil {
		return
	}
	wants := false
	if focused != nil {
		if wi, ok := any(focused).(WantsIME); ok && wi.WantsIME() {
			wants = true
		}
	}
	if wants {
		c.ime.Show()
	} else {
		c.ime.Hide()
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
		if c.widgets[idx].IsVisible() && c.widgets[idx].IsEnabled() && c.widgets[idx].Focusable() {
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
		if c.widgets[idx].IsVisible() && c.widgets[idx].IsEnabled() && c.widgets[idx].Focusable() {
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
		} else {
			c.ptr.IsDown = false
			c.ptr.IsTouch = true
			c.ptr.IsJustUp = true
			c.hasTouch = false
		}

		return
	}

	c.ptr.Position.X, c.ptr.Position.Y = ebiten.CursorPosition()
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
		if !w.IsVisible() || !w.IsEnabled() {
			continue
		}

		if c.widgetHit(w, pos) {
			return w
		}
	}

	return nil
}

func (c *Context) Update() {
	c.readPointerSnapshot()
	c.root.SetHeight(c.dstBounds.Dy())
	c.root.SetFrame(0, 0, c.dstBounds.Dx())
	c.root.Update(c)

	c.rebuildWidgets()

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

	for _, w := range c.widgets {
		if !w.IsVisible() {
			continue
		}

		w.SetHovered(hoverTarget == w)

		// Pointer down routed to the chosen target.
		if c.ptr.IsJustDown && target == w && w.IsEnabled() {
			w.SetPressed(true)
			w.Dispatch(Event{Widget: w, Type: EventPointerDown, Pointer: c.ptr})
		}

		// Pointer up: release + click if pointer ends inside widget.
		if c.ptr.IsJustUp {
			wasPressed := w.IsPressed()
			if wasPressed {
				w.Dispatch(Event{Widget: w, Type: EventPointerUp, Pointer: c.ptr})

				if w.IsEnabled() && c.widgetHit(w, c.ptr.Position) {
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
	c.root.Draw(c, dst)
	c.root.DrawOverlay(c, dst)
}
