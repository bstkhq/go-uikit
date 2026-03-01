package widget

import (
	"image/color"

	"github.com/bstkhq/go-uikit"
	"github.com/tinne26/etxt"
	"golang.org/x/image/font/sfnt"
)

type TextModifier func(*uikit.Theme, *etxt.Renderer)

func Font(f *sfnt.Font) TextModifier {
	return func(theme *uikit.Theme, renderer *etxt.Renderer) {
		renderer.SetFont(f)
	}
}

func Color(c color.Color) TextModifier {
	return func(theme *uikit.Theme, renderer *etxt.Renderer) {
		renderer.SetColor(c)
	}
}

func Size(s float64) TextModifier {
	return func(theme *uikit.Theme, renderer *etxt.Renderer) {
		renderer.SetSize(s)
	}
}
