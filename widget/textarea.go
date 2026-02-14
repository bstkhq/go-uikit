package widget

import (
	"math"
	"strings"

	"github.com/erparts/go-uikit"
	"github.com/erparts/go-uikit/common"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
)

var _ uikit.Widget = (*TextArea)(nil)

// TextArea is a multi-line text editor with internal vertical scrolling.
type TextArea struct {
	uikit.Base

	text        string
	placeholder string

	lines  int
	Scroll uikit.Scroller

	// Caret config (end-of-text caret)
	CaretWidthPx  int
	CaretBlinkMs  int
	CaretMarginPx int
	caretTick     int

	// Reusable buffers (avoid allocations every frame)
	inputBuf  []rune
	appendBuf []rune
}

func NewTextArea(theme *uikit.Theme, placeholder string) *TextArea {
	cfg := uikit.NewWidgetBaseConfig(theme)

	w := &TextArea{
		Base:          uikit.NewBase(cfg),
		placeholder:   placeholder,
		lines:         5,
		CaretWidthPx:  2,
		CaretBlinkMs:  600,
		CaretMarginPx: 0,
	}

	w.Scroll = uikit.NewScroller()
	w.Scroll.Scrollbar = uikit.ScrollbarAlways

	w.Base.HeightCalculator = w.calculateHeight
	return w
}

func (w *TextArea) calculateHeight() int {
	lines := w.lines
	if lines <= 0 {
		lines = 5
	}

	theme := w.Theme()
	lineH := int(math.Ceil(theme.Text().Utils().GetLineHeight()))
	controlH := theme.PadY*2 + lines*lineH
	if controlH < theme.ControlH {
		controlH = theme.ControlH
	}
	return controlH
}

func (w *TextArea) Focusable() bool { return true }
func (w *TextArea) WantsIME() bool  { return true }
func (w *TextArea) Text() string    { return w.text }

func (w *TextArea) SetText(s string) {
	if w.text == s {
		return
	}
	w.text = s
	w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventValueChange})
}

func (w *TextArea) SetLines(n int) {
	if n < 1 {
		n = 1
	}
	w.lines = n
}

// setTextInternal updates the text without multiple dispatches.
// Caller decides whether to Dispatch once.
func (w *TextArea) setTextInternal(s string) {
	w.text = s
}

func (w *TextArea) Update(ctx *uikit.Context) {
	focused := w.IsFocused()
	enabled := w.IsEnabled()

	if focused && enabled {
		w.caretTick++
	} else {
		w.caretTick = 0
	}

	theme := ctx.Theme()
	content := common.Inset(w.Measure(false), theme.PadX, theme.PadY)

	// Only line-height is needed for scroll math.
	t := theme.Text()
	lineH := int(math.Ceil(t.Utils().GetLineHeight()))
	if lineH <= 0 {
		lineH = 1
	}

	// Compute line count without splitting
	lineCount := 1
	if w.text != "" {
		lineCount = 1 + strings.Count(w.text, "\n")
	}

	contentH := lineCount * lineH
	if contentH < content.Dy() {
		contentH = content.Dy()
	}

	w.Scroll.Update(ctx, content, contentH)

	if !focused || !enabled {
		return
	}

	original := w.text
	text := original

	// --- IME / chars (buffer reuse) ---
	w.inputBuf = ebiten.AppendInputChars(w.inputBuf[:0])
	w.appendBuf = w.appendBuf[:0]

	flushAppend := func() {
		if len(w.appendBuf) == 0 {
			return
		}
		text += string(w.appendBuf)
		w.appendBuf = w.appendBuf[:0]
	}

	changed := false

	for _, ch := range w.inputBuf {
		// backspace can come as '\b' or DEL
		if ch == '\b' || ch == 0x7f {
			flushAppend()
			text = removeLastRune(text)
			changed = true
			continue
		}

		// newline
		if ch == '\n' || ch == '\r' {
			flushAppend()
			text += "\n"
			changed = true
			continue
		}

		// control chars ignored
		if ch < 0x20 {
			continue
		}

		w.appendBuf = append(w.appendBuf, ch)
	}

	if len(w.appendBuf) > 0 {
		flushAppend()
		changed = true
	}

	// Fallback key handling for platforms that don't deliver via AppendInputChars
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		text = removeLastRune(text)
		changed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		text += "\n"
		changed = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ctx.SetFocus(nil)
	}

	// Apply once + Dispatch once
	if changed && text != original {
		w.setTextInternal(text)
		w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventValueChange})

		// Scroll-to-caret (caret at end of text)
		newLineCount := 1 + strings.Count(w.text, "\n")
		caretBottom := newLineCount * lineH

		minScroll := caretBottom - content.Dy()
		if minScroll < 0 {
			minScroll = 0
		}
		if w.Scroll.ScrollY < minScroll {
			w.Scroll.ScrollY = minScroll
		}

		maxScroll := (newLineCount*lineH - content.Dy())
		if maxScroll < 0 {
			maxScroll = 0
		}
		if w.Scroll.ScrollY > maxScroll {
			w.Scroll.ScrollY = maxScroll
		}
	}
}

func (w *TextArea) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	r := w.Base.Draw(ctx, dst)

	theme := ctx.Theme()
	content := common.Inset(r, theme.PadX, theme.PadY)

	// Clip to content
	sub := dst.SubImage(content).(*ebiten.Image)
	ox, oy := content.Bounds().Min.X, content.Bounds().Min.Y

	t := theme.Text()
	t.SetFont(theme.Font)
	t.SetSize(float64(theme.FontPx))
	t.SetAlign(etxt.Left | etxt.Top)

	lineH := t.Measure(" ").IntHeight()
	if lineH <= 0 {
		lineH = 1
	}

	startY := -w.Scroll.ScrollY

	// Placeholder
	drawStr := w.text
	col := theme.TextColor
	if drawStr == "" && !w.IsFocused() {
		drawStr = w.placeholder
		col = theme.MutedTextColor
	}

	t.SetColor(col)

	// Draw lines WITHOUT strings.Split allocation
	// Iterate and draw between '\n' boundaries.
	y := startY
	from := 0
	for i := 0; i <= len(drawStr); i++ {
		end := i
		if i == len(drawStr) || drawStr[i] == '\n' {
			line := drawStr[from:end]
			t.Draw(sub, line, ox, oy+y)
			y += lineH
			from = i + 1
		}
	}

	// Scrollbar
	lineCount := 1
	if w.text != "" {
		lineCount = 1 + strings.Count(w.text, "\n")
	}
	contentH := lineCount * lineH
	if contentH < content.Dy() {
		contentH = content.Dy()
	}

	w.Scroll.DrawBar(sub, theme, content.Dx(), content.Dy(), contentH)

	// Caret at end
	if w.IsFocused() && w.IsEnabled() && w.CaretWidthPx > 0 {
		blinkFrames := int(math.Max(1, float64(w.CaretBlinkMs)/1000.0*60.0))
		if (w.caretTick/blinkFrames)%2 == 0 {

			// last line text = substring after last '\n'
			lastIdx := strings.LastIndex(w.text, "\n")
			lastLine := w.text
			lineIdx := 0
			if lastIdx >= 0 {
				lastLine = w.text[lastIdx+1:]
				lineIdx = 1 + strings.Count(w.text[:lastIdx], "\n")
			}

			wBefore := t.Measure(lastLine).ImageRect().Dx()

			cx := wBefore + w.CaretMarginPx
			cy := (lineIdx * lineH) - w.Scroll.ScrollY

			if cx < 0 {
				cx = 0
			}
			maxX := content.Dx() - w.CaretWidthPx
			if maxX < 0 {
				maxX = 0
			}
			if cx > maxX {
				cx = maxX
			}

			if cy+lineH > 0 && cy <= content.Dy() {
				vector.DrawFilledRect(
					sub,
					float32(ox+cx),
					float32(oy+cy),
					float32(w.CaretWidthPx),
					float32(lineH),
					theme.CaretColor,
					false,
				)
			}
		}
	}
}
