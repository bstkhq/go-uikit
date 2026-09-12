package uikit

import (
	"testing"

	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"
)

func TestNewThemeUsesRegularAsDefaultBold(t *testing.T) {
	regular := parseTestFont(t, goregular.TTF)
	theme := NewTheme(regular, 20)

	if theme.Font != regular || theme.BoldFont != regular {
		t.Fatal("NewTheme must use its font for both regular and bold text")
	}
}

func TestThemeFontsAreIndependentAndPreserveConfiguration(t *testing.T) {
	regular := parseTestFont(t, goregular.TTF)
	bold := parseTestFont(t, gobold.TTF)
	theme := NewThemeWithFonts(regular, bold, 24)
	theme.CaretWidthPx = 9
	controlHeight := theme.ControlH
	textColor := theme.TextColor

	replacementRegular := parseTestFont(t, gobold.TTF)
	replacementBold := parseTestFont(t, goregular.TTF)
	theme.SetFonts(replacementRegular, replacementBold)

	if theme.Font != replacementRegular || theme.BoldFont != replacementBold {
		t.Fatal("SetFonts did not update both font pointers")
	}
	if theme.ControlH != controlHeight || theme.TextColor != textColor || theme.CaretWidthPx != 9 {
		t.Fatal("SetFonts changed non-font theme configuration")
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
