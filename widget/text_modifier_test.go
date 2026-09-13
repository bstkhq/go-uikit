package widget

import (
	"image/color"
	"testing"

	"github.com/bstkhq/go-uikit"
)

func TestTextStyleSelectsThemeFontOnly(t *testing.T) {
	theme := uikit.DefaultTheme()
	renderer := theme.Text()
	renderer.SetSize(31)
	renderer.SetColor(color.RGBA{1, 2, 3, 4})

	TextStyle(uikit.TextBold)(theme, renderer)

	if renderer.GetFont() != theme.Font(uikit.TextBold) {
		t.Fatal("TextStyle did not select the requested theme font")
	}
	if renderer.GetSize() != 31 {
		t.Fatal("TextStyle changed the renderer size")
	}
	if renderer.GetColor() != (color.RGBA{1, 2, 3, 4}) {
		t.Fatal("TextStyle changed the renderer color")
	}
}
