package widget

import (
	"math"
	"time"
	"unicode/utf8"

	"github.com/erparts/go-uikit"
	"github.com/erparts/go-uikit/common"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var _ uikit.Widget = (*TextInput)(nil)

// TextInput is a single-line input box (no label).
// Height and proportions come from Theme; external layout controls only width.
type TextInput struct {
	uikit.Base

	text        string
	placeholder string
	caretTick   int

	// Reusable buffers to avoid allocations on every Update().
	inputBuf  []rune
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

func (w *TextInput) Focusable() bool { return true }
func (w *TextInput) WantsIME() bool  { return true }
func (w *TextInput) Text() string    { return w.text }

// SetText sets the current text value and dispatches a value-change event.
func (w *TextInput) SetText(s string) {
	if w.text == s {
		return
	}
	w.text = s
	w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventValueChange})
}

// SetTextSilently sets the current text value without dispatching events.
// Useful internally to batch changes and dispatch once.
func (w *TextInput) SetTextSilently(s string) {
	w.text = s
}

// AppendText appends a string to the current text and dispatches a value-change event.
func (w *TextInput) AppendText(s string) {
	if s == "" {
		return
	}
	w.SetText(w.text + s)
}

// Reset clears the current text.
func (w *TextInput) Reset() {
	w.SetText("")
}

// removeLastRune removes the last UTF-8 rune from the provided string.
func removeLastRune(s string) string {
	if s == "" {
		return ""
	}
	_, sz := utf8.DecodeLastRuneInString(s)
	if sz <= 0 || sz > len(s) {
		return ""
	}
	return s[:len(s)-sz]
}

func (w *TextInput) Update(ctx *uikit.Context) {
	focused := w.IsFocused()
	enabled := w.IsEnabled()

	if focused && enabled {
		w.caretTick++
	} else {
		w.caretTick = 0
	}

	if !focused || !enabled {
		return
	}

	original := w.text
	text := original

	// Reuse buffer to avoid allocations.
	w.inputBuf = ebiten.AppendInputChars(w.inputBuf[:0])

	// Batch normal runes to avoid repeated string concatenations.
	w.appendBuf = w.appendBuf[:0]

	flushAppend := func() {
		if len(w.appendBuf) == 0 {
			return
		}
		text += string(w.appendBuf)
		w.appendBuf = w.appendBuf[:0]
	}

	// IME / input chars
	for _, ch := range w.inputBuf {
		// Backspace can arrive as '\b' or DEL.
		if ch == '\b' || ch == 0x7f {
			flushAppend()
			text = removeLastRune(text)
			continue
		}

		// Skip control characters.
		if ch < 0x20 {
			continue
		}

		w.appendBuf = append(w.appendBuf, ch)
	}

	flushAppend()

	// Desktop / fallback backspace handling (Android IME can be inconsistent).
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		text = removeLastRune(text)
	}

	// Commit focus changes (no text modification).
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		ctx.SetFocus(nil)
	}

	// Dispatch only once if something actually changed.
	if text != original {
		w.SetText(text)
	}
}

func (w *TextInput) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	w.Base.Draw(ctx, dst)

	theme := ctx.Theme()
	r := w.Measure(false)

	content := common.Inset(r, theme.PadX, theme.PadY)
	middleY := r.Min.Y + r.Dy()/2

	// Decide what to render: actual text or placeholder.
	drawStr := w.text
	textCol := theme.TextColor
	if drawStr == "" && !w.IsFocused() {
		drawStr = w.placeholder
		textCol = theme.MutedTextColor
	}

	t := theme.Text()

	// Horizontal overflow handling (keep the end visible).
	m := t.Measure(drawStr)
	textW := m.ImageRect().Dx()

	shiftX := 0
	if textW > content.Dx() {
		shiftX = content.Dx() - textW
	}

	// Draw text centered vertically.
	t.SetColor(textCol)
	t.Draw(dst, drawStr, content.Min.X+shiftX, middleY)

	// Caret drawing (end-of-text caret).
	if w.IsFocused() && w.IsEnabled() && theme.CaretWidthPx > 0 {
		blinkFrames := int(math.Max(1, float64(theme.CaretBlink)/float64(time.Second)*60.0))
		if (w.caretTick/blinkFrames)%2 == 0 {
			measureStr := w.text
			if measureStr == "" {
				measureStr = " "
			}

			mc := t.Measure(measureStr)

			cx := content.Min.X + shiftX + mc.IntWidth() + theme.CaretMarginPx
			if w.text == "" {
				cx = content.Min.X + shiftX + theme.CaretMarginPx
			}

			caretH := mc.IntHeight()
			cy := middleY - (caretH / 2)

			// Clamp caret into content rect.
			if cx < content.Min.X {
				cx = content.Min.X
			}
			if cx > content.Min.X+content.Dx() {
				cx = content.Min.X + content.Dx()
			}

			vector.DrawFilledRect(
				dst,
				float32(cx),
				float32(cy),
				float32(theme.CaretWidthPx),
				float32(caretH),
				theme.CaretColor,
				false,
			)
		}
	}
}
