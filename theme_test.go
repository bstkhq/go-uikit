package uikit

import (
	"image/color"
	"testing"

	"github.com/tinne26/etxt"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"
)

func TestNewThemeRegistersDefaultStyle(t *testing.T) {
	regular := parseTestFont(t, goregular.TTF)
	theme := NewTheme(regular, 24)

	if theme.TextStyles[TextDefault] != regular {
		t.Fatal("NewTheme did not register its font as TextDefault")
	}
	if theme.FontPx != 24 || theme.ControlH <= 0 {
		t.Fatal("NewTheme did not preserve its size-derived configuration")
	}
}

func TestDefaultThemeProvidesRegularAndBoldStyles(t *testing.T) {
	theme := DefaultTheme()
	regular := theme.TextStyles[TextDefault]
	bold := theme.TextStyles[TextBold]
	if regular == nil || bold == nil {
		t.Fatal("DefaultTheme must provide regular and bold font styles")
	}
	if regular == bold {
		t.Fatal("DefaultTheme must use distinct regular and bold font faces")
	}
}

func TestRendererUsesRequestedStyleAndDefaults(t *testing.T) {
	theme := DefaultTheme()
	theme.FontPx = 27
	theme.TextColor = color.RGBA{1, 2, 3, 4}

	renderer := theme.Renderer(TextBold)
	if renderer.GetFont() != theme.TextStyles[TextBold] {
		t.Fatal("Renderer did not select the requested style")
	}
	if renderer.Fract().GetScaledSize().ToFloat64() != 27 {
		t.Fatal("Renderer did not apply Theme.FontPx")
	}
	if renderer.GetColor() != theme.TextColor {
		t.Fatal("Renderer did not apply Theme.TextColor")
	}
	if renderer.GetAlign() != etxt.Left|etxt.VertCenter {
		t.Fatal("Renderer did not restore the default alignment")
	}

	italicStyle := TextStyle("italic")
	italic := parseTestFont(t, goregular.TTF)
	theme.TextStyles[italicStyle] = italic
	if theme.Renderer(italicStyle).GetFont() != italic {
		t.Fatal("Renderer did not select an application-defined style")
	}

	if theme.Renderer(TextStyle("missing")).GetFont() != theme.TextStyles[TextDefault] {
		t.Fatal("missing styles must fall back to TextDefault")
	}
	theme.TextStyles[TextBold] = nil
	if theme.Renderer(TextBold).GetFont() != theme.TextStyles[TextDefault] {
		t.Fatal("nil styles must fall back to TextDefault")
	}
}

func TestFontResolvesStylesWithoutMutatingRenderer(t *testing.T) {
	theme := DefaultTheme()
	renderer := theme.Renderer(TextDefault)
	defaultFont := renderer.GetFont()

	if theme.Font(TextBold) != theme.TextStyles[TextBold] {
		t.Fatal("Font did not resolve the requested style")
	}
	if renderer.GetFont() != defaultFont {
		t.Fatal("Font must not mutate the shared renderer")
	}
	if theme.Font(TextStyle("missing")) != defaultFont {
		t.Fatal("Font must fall back to TextDefault")
	}
}

func TestTextUsesDefaultStyle(t *testing.T) {
	theme := DefaultTheme()
	if theme.Text().GetFont() != theme.TextStyles[TextDefault] {
		t.Fatal("Text must use TextDefault")
	}
	if theme.Text() != theme.Renderer(TextDefault) {
		t.Fatal("Text and Renderer must share the theme renderer")
	}
}

func parseTestFont(t *testing.T, data []byte) *sfnt.Font {
	t.Helper()
	font, err := sfnt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return font
}
