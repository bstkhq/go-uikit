package widget

import (
	"image"
	"math"
	"slices"

	"github.com/bstkhq/go-uikit"
	"github.com/bstkhq/go-uikit/common"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/fract"
)

var _ uikit.Widget = (*TextInput)(nil)

// TextInput is a single-line input box (no label).
// Height and proportions come from Theme; external layout controls only width.
type TextInput struct {
	uikit.Base

	text        string
	placeholder string
	caretTick   int
	caretPos    int
	scrollPos   int  // in runes
	anchorRight bool // for scrollPos
	wasFocused  bool

	IMEOptions uikit.IMEOptions
	OnCommit   func(*uikit.Context)

	inputRuneLimit int
	inputLimitTick int

	// Reusable buffers to avoid allocations on every Update().
	textRunes []rune
	appendBuf []rune
}

func NewTextInput(theme *uikit.Theme, placeholder string) *TextInput {
	cfg := uikit.NewWidgetBaseConfig(theme)

	w := &TextInput{
		placeholder: placeholder,
	}

	w.Base = uikit.NewBase(cfg)
	return w
}

func (w *TextInput) Focusable() bool       { return true }
func (w *TextInput) IME() uikit.IMEOptions { return w.IMEOptions }
func (w *TextInput) Text() string          { return w.text }

// SetText sets the current text value and dispatches a value-change event.
func (w *TextInput) SetText(s string) {
	if w.text == s {
		return
	}

	w.SetTextSilently(s)
	w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventValueChange})
}

// SetTextSilently sets the current text value without dispatching events.
// Useful internally to batch changes and dispatch once.
func (w *TextInput) SetTextSilently(s string) {
	if w.text == s {
		return
	}

	w.text = s
	w.textRunes = w.textRunes[:0]
	for _, r := range s {
		w.textRunes = append(w.textRunes, r)
	}
	w.caretPos = min(w.caretPos, len(w.textRunes))
	w.scrollPos = max(0, w.scrollPos)
}

// SetInputRuneLimit limits the max number of code points that can be
// entered by the user.
//
// Notice that this doesn't apply retroactively nor to manual operations
// like [TextInput.SetText]().
//
// If limit <= 0, the input is unlimited.
func (w *TextInput) SetInputRuneLimit(limit int) {
	w.inputRuneLimit = max(limit, 0)
}

// Caret returns the current caret index, in runes.
func (w *TextInput) Caret() int {
	return w.caretPos
}

// SetCaret manually changes the caret position, in runes.
// Values out of range will be clamped.
func (w *TextInput) SetCaret(index int) {
	w.caretPos = min(max(index, 0), len(w.textRunes))
}

// ClosestCaretIndex performs hit testing to find the caret index
// closest to the given X coordinate. Since TextInput only has one
// line of text, the Y coordinate is not required.
func (w *TextInput) ClosestCaretIndex(x int) int {
	theme := w.Theme()
	r := w.Measure(false)
	r = r.Inset(theme.PadX)

	if x <= r.Min.X {
		return 0
	}
	if x >= r.Max.X {
		return len(w.textRunes)
	}

	// find shift
	feed := etxt.NewFeed(theme.Text())
	for i, r := range w.textRunes {
		if i == w.scrollPos {
			break
		}
		feed.Advance(r)
	}
	shift := -feed.Position.X
	if w.anchorRight {
		shift += fract.FromInt(r.Dx())
	}

	// find nearest caret
	feed.Reset()
	feed.Renderer = theme.Text()
	fx := fract.FromInt(x)
	feed.Position.X = fract.FromInt(r.Min.X) + shift
	closestDist := (fx - feed.Position.X).Abs()
	for i, r := range w.textRunes {
		feed.Advance(r)
		dist := (fx - feed.Position.X).Abs()
		if dist > closestDist { // prev was best
			return i
		}
		closestDist = dist
	}
	return len(w.textRunes)
}

// Reset clears the current text.
func (w *TextInput) Reset() {
	w.wasFocused = false
	w.SetText("")
}

func (w *TextInput) Update(ctx *uikit.Context) {
	w.caretTick += 1
	if w.wasFocused != w.IsFocused() {
		if !w.wasFocused {
			w.caretTick = 0
			w.wasFocused = !w.wasFocused
		}
	}
	if w.inputLimitTick > 0 {
		w.inputLimitTick += 1
		if w.inputLimitTick > (ebiten.TPS()*2)/5 {
			w.inputLimitTick = 0
		}
	}

	if !w.IsEnabled() {
		return
	}

	ptr := ctx.Pointer()
	if ptr.IsDown && ptr.Position.In(w.Measure(false)) {
		w.caretTick = 0
		index := w.ClosestCaretIndex(ptr.Position.X)
		w.SetCaret(index)
	}

	if !w.IsFocused() {
		return
	}

	var changed bool
	w.appendBuf = ebiten.AppendInputChars(w.appendBuf[:0])
	for _, r := range w.appendBuf {
		switch {
		case r == 0x08: // BS
			changed = w.deleteRuneBS() || changed
		case r == 0x7f: // DEL
			changed = w.deleteRuneDEL() || changed
		case r >= 0x20: // append runes >= ' ' (ignore control chars)
			if w.inputRuneLimit == 0 || len(w.textRunes) < w.inputRuneLimit {
				w.textRunes = slices.Insert(w.textRunes, w.caretPos, r) // could be optimized
				w.caretPos += 1
				changed = true
			} else {
				w.caretTick = 0
				w.inputLimitTick = 1
			}
		}
	}

	// handle other special keys
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		changed = w.deleteRuneBS() || changed
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		changed = w.deleteRuneDEL() || changed
	}
	w.updateNav()

	// update text and dispatch changes
	if changed {
		w.caretTick = 0
		w.text = string(w.textRunes)
		w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventValueChange})
	}

	// remove focus on Enter
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		if w.OnCommit != nil {
			w.OnCommit(ctx)
		} else {
			ctx.SetFocus(nil)
		}
	}
}

func (w *TextInput) updateNav() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		w.caretTick = 0
		if w.caretPos > 0 {
			if ebiten.IsKeyPressed(ebiten.KeyControl) {
				w.caretPos = prevBreakPos(w.textRunes, w.caretPos)
			} else {
				w.caretPos -= 1
			}
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		w.caretTick = 0
		if w.caretPos < len(w.textRunes) {
			if ebiten.IsKeyPressed(ebiten.KeyControl) {
				w.caretPos = nextBreakPos(w.textRunes, w.caretPos)
			} else {
				w.caretPos += 1
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) || inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		w.caretPos = 0
		w.caretTick = 0
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEnd) || inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		w.caretPos = len(w.textRunes)
		w.caretTick = 0
	}
}

func (w *TextInput) deleteRuneBS() bool {
	if len(w.textRunes) == 0 || w.caretPos <= 0 {
		return false
	}
	start := w.caretPos - 1
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		start = prevBreakPos(w.textRunes, w.caretPos)
	}
	w.textRunes = slices.Delete(w.textRunes, start, w.caretPos)
	w.caretPos -= (w.caretPos - start)
	return true
}

func (w *TextInput) deleteRuneDEL() bool {
	if len(w.textRunes) == 0 {
		return false
	}
	if w.caretPos >= len(w.textRunes) {
		return w.deleteRuneBS()
	}

	end := w.caretPos + 1
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		end = nextBreakPos(w.textRunes, w.caretPos)
	}

	w.textRunes = slices.Delete(w.textRunes, w.caretPos, end)
	return true
}

func (w *TextInput) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	r := w.Base.Draw(ctx, dst)
	theme := ctx.Theme()
	cy := r.Min.Y + r.Dy()/2
	r = common.Inset(r, theme.PadX, theme.PadY)

	// if no text and unfocused, draw placeholder
	renderer := theme.Text()
	if len(w.textRunes) == 0 && !w.IsFocused() {
		if w.placeholder != "" {
			renderer.SetColor(theme.MutedTextColor)
			renderer.Draw(dst, w.placeholder, r.Min.X, cy)
		}
		return
	}

	// find caret and scroll anchor positions
	feed := etxt.NewFeed(renderer)
	caretShift, scrollShift := -1, -1
	for i, r := range w.textRunes {
		if i == w.caretPos {
			caretShift = feed.Position.X.ToIntFloor()
			if scrollShift != -1 {
				break
			}
		}
		if i == w.scrollPos {
			scrollShift = feed.Position.X.ToIntFloor()
			if caretShift != -1 {
				break
			}
		}
		feed.Advance(r)
	}
	if caretShift == -1 {
		caretShift = feed.Position.X.ToIntFloor()
	}
	if scrollShift == -1 {
		scrollShift = feed.Position.X.ToIntFloor()
	}

	// adjust scroll to view area
	width := r.Dx()
	if w.anchorRight {
		if scrollShift-caretShift > width { // switch to anchor left
			w.anchorRight = false
			w.scrollPos = w.caretPos
			scrollShift = caretShift
		} else if feed.Position.X.ToIntFloor() < width { // restore full left anchor
			w.anchorRight = false
			w.scrollPos = 0
			scrollShift = 0
		} else if w.caretPos > w.scrollPos { // expand right
			w.scrollPos = w.caretPos
			scrollShift = caretShift
		}
	} else { // anchor left
		if caretShift-scrollShift > width { // switch to anchor right
			w.anchorRight = true
			w.scrollPos = w.caretPos
			scrollShift = caretShift
		} else if w.caretPos < w.scrollPos { // expand left
			w.scrollPos = w.caretPos
			scrollShift = caretShift
		}
	}

	// draw text
	shift := -scrollShift
	if w.anchorRight {
		shift += width
	}
	renderer.Draw(dst, w.text, r.Min.X+shift, cy)

	// draw caret
	if w.IsFocused() && w.IsEnabled() && theme.CaretWidthPx > 0 && w.blink(theme) {
		lineHeight := int(math.Round(renderer.Utils().GetLineHeight()))
		x := r.Min.X + caretShift + shift + theme.CaretMarginPx
		cy -= (lineHeight / 2)
		b := dst.Bounds()
		cy -= b.Min.Y
		x -= b.Min.X
		clr := theme.CaretColor
		if w.inputLimitTick > 0 {
			clr = theme.ErrorTextColor
		}
		w.Base.DrawRoundedRect(dst, image.Rect(x, cy, x+theme.CaretWidthPx, cy+lineHeight), 0, clr)
	}
}

func (w *TextInput) blink(theme *uikit.Theme) bool {
	ticks := blinkTicks(theme)
	if ticks <= 0 {
		return false
	}
	return (w.caretTick/ticks)%2 == 0
}
