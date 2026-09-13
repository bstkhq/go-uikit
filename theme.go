package uikit

import (
	"image/color"
	"math"
	"time"

	"github.com/tinne26/etxt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// Theme is the single source of truth for widget proportions.
// Only the font and font size are inputs; everything else derives from them.
type Theme struct {
	TextStyles TextStyles
	FontPx     int
	ControlH   int

	PadX int
	PadY int

	Radius       int
	BorderW      int
	FocusRingW   int
	FocusRingGap int

	SpaceS int
	SpaceM int
	SpaceL int

	CheckSize int // checkbox square size

	// Validation
	ErrorFontPx int
	ErrorGap    int

	// Colors
	TextColor           color.RGBA
	MutedTextColor      color.RGBA
	BackgroundColor     color.RGBA
	SurfaceColor        color.RGBA
	SurfaceHoverColor   color.RGBA
	SurfacePressedColor color.RGBA
	BorderColor         color.RGBA
	FocusColor          color.RGBA
	DisabledColor       color.RGBA
	ErrorTextColor      color.RGBA
	ErrorBorderColor    color.RGBA
	Scrollbar           color.RGBA
	CaretColor          color.RGBA

	// Scrollbar
	ScrollbarRadius int
	ScrollbarMinH   int
	ScrollbarW      int

	// Caret
	CaretWidthPx  int
	CaretBlink    time.Duration
	CaretMarginPx int

	renderer *etxt.Renderer
}

// TextStyle identifies a font variant within a theme. Applications may define
// additional variants, such as "italic" or "narrow".
type TextStyle string

const (
	TextDefault TextStyle = "default"
	TextBold    TextStyle = "bold"
)

// TextStyles maps font variants to their faces. Font size is shared by all
// variants through [Theme.FontPx].
type TextStyles map[TextStyle]*sfnt.Font

// Font returns the font registered for style. Missing or nil variants fall
// back to TextDefault.
func (t *Theme) Font(style TextStyle) *sfnt.Font {
	font := t.TextStyles[style]
	if font == nil {
		font = t.TextStyles[TextDefault]
	}
	return font
}

// Renderer returns the shared text renderer configured for style. Missing or
// nil variants fall back to TextDefault.
func (t *Theme) Renderer(style TextStyle) *etxt.Renderer {
	if t.renderer == nil {
		r := etxt.NewRenderer()
		r.Utils().SetCache8MiB()
		r.Glyph().SetMissHandler(etxt.OnMissNotdef)
		t.renderer = r
	}

	t.renderer.SetFont(t.Font(style))
	t.renderer.SetSize(float64(t.FontPx))
	t.renderer.SetColor(t.TextColor)
	t.renderer.SetAlign(etxt.Left | etxt.VertCenter)
	return t.renderer
}

// Text returns the shared renderer configured with the default text style.
func (t *Theme) Text() *etxt.Renderer {
	return t.Renderer(TextDefault)
}

func (t *Theme) ErrorText() *etxt.Renderer {
	r := t.Text()
	r.SetSize(float64(t.ErrorFontPx))
	r.SetColor(t.ErrorTextColor)
	return r
}

func fontHeight(f *sfnt.Font, sizePx int) (int, error) {
	var buf sfnt.Buffer
	m, err := f.Metrics(&buf, fixed.I(sizePx), font.HintingNone)
	if err != nil {
		return 0, err
	}

	return m.Height.Round(), nil
}

func DefaultTheme() *Theme {
	regular, _ := sfnt.Parse(goregular.TTF)
	bold, _ := sfnt.Parse(gobold.TTF)
	theme := NewTheme(regular, 20)
	theme.TextStyles[TextBold] = bold
	return theme
}

func NewTheme(font *sfnt.Font, fontPx int) *Theme {
	if fontPx < 10 {
		fontPx = 10
	}

	fontH, err := fontHeight(font, fontPx)
	if err != nil {
		panic(err)
	}

	// Control height derived from font height.
	// k=1.8 gives a "web input" feel without being too tall.
	controlH := int(math.Round(float64(fontH) * 1.8))
	if controlH < fontH+6 {
		controlH = fontH + 6
	}

	padY := (controlH - fontH) / 2
	if padY < 2 {
		padY = 2
	}
	padX := int(math.Round(float64(padY) * 1.6))
	if padX < 6 {
		padX = 6
	}

	radius := int(math.Round(float64(controlH) * 0.22))
	if radius < 4 {
		radius = 4
	}

	borderW := int(math.Round(float64(controlH) * 0.06))
	if borderW < 1 {
		borderW = 1
	}

	focusW := borderW
	focusGap := int(math.Round(float64(borderW) * 1.5))
	if focusGap < 2 {
		focusGap = 2
	}

	spaceS := int(math.Round(float64(controlH) * 0.20))
	spaceM := int(math.Round(float64(controlH) * 0.35))
	spaceL := int(math.Round(float64(controlH) * 0.60))
	if spaceS < 4 {
		spaceS = 4
	}
	if spaceM < 8 {
		spaceM = 8
	}
	if spaceL < 12 {
		spaceL = 12
	}

	checkSize := fontH
	if checkSize < controlH-padY*2 {
		checkSize = controlH - padY*2
	}

	// Keep it visually balanced
	checkSize = int(math.Round(float64(checkSize) * 0.92))

	// Validation
	errorFontPx := int(math.Round(float64(fontPx) * 0.85))
	if errorFontPx < 10 {
		errorFontPx = 10
	}

	errorGap := int(math.Round(float64(controlH) * 0.15))
	if errorGap < 4 {
		errorGap = 4
	}

	return &Theme{
		TextStyles: TextStyles{TextDefault: font},
		FontPx:     fontPx,
		ControlH:   controlH,
		PadX:       padX,
		PadY:       padY,

		Radius:       radius,
		BorderW:      borderW,
		FocusRingW:   focusW,
		FocusRingGap: focusGap,

		SpaceS: spaceS,
		SpaceM: spaceM,
		SpaceL: spaceL,

		CheckSize: checkSize,

		ErrorFontPx: errorFontPx,
		ErrorGap:    errorGap,

		TextColor:           color.RGBA{235, 238, 242, 255},
		MutedTextColor:      color.RGBA{170, 176, 186, 255},
		BackgroundColor:     color.RGBA{20, 22, 26, 255},
		SurfaceColor:        color.RGBA{34, 38, 46, 255},
		SurfaceHoverColor:   color.RGBA{42, 48, 58, 255},
		SurfacePressedColor: color.RGBA{28, 32, 40, 255},
		BorderColor:         color.RGBA{76, 84, 98, 255},
		FocusColor:          color.RGBA{120, 170, 255, 255},
		DisabledColor:       color.RGBA{90, 96, 106, 255},
		ErrorTextColor:      color.RGBA{235, 110, 110, 255},
		ErrorBorderColor:    color.RGBA{235, 110, 110, 255},

		CaretColor:    color.RGBA{235, 238, 242, 255},
		CaretWidthPx:  2,
		CaretBlink:    600 * time.Millisecond,
		CaretMarginPx: 0,
	}
}
