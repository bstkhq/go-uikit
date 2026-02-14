package uikit

import (
	"image"
	"image/color"

	"github.com/erparts/go-shapes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
)

var shapesRenderer = shapes.NewRenderer()

func clampInt(v, lo, hi int) int {
	return max(lo, min(hi, v))
}

func drawPathOptionsForColor(col color.RGBA) *vector.DrawPathOptions {
	op := &vector.DrawPathOptions{
		AntiAlias: true,
	}
	op.ColorScale.ScaleWithColor(col)
	return op
}

func drawRoundedRect(dst *ebiten.Image, r image.Rectangle, radius int, col color.RGBA) {
	if r.Empty() {
		return
	}

	rr := clampInt(radius, 0, min(r.Dx()/2, r.Dy()/2))
	shapesRenderer.SetColor(col)
	shapesRenderer.DrawRect(dst, r, float32(rr))
}

func drawRoundedBorder(dst *ebiten.Image, r image.Rectangle, radius int, borderW int, col color.RGBA) {
	if borderW <= 0 || r.Empty() {
		return
	}

	rr := clampInt(radius, 0, min(r.Dx()/2, r.Dy()/2))
	bw := float32(borderW)

	shapesRenderer.SetColor(col)
	shapesRenderer.StrokeRect(dst, r, 0, bw, float32(rr))
}

func drawErrorText(ctx *Context, dst *ebiten.Image, r image.Rectangle, msg string) {
	if msg == "" || r.Empty() {
		return
	}

	t := ctx.Theme().ErrorText()
	t.SetAlign(etxt.Left | etxt.Top)
	t.Draw(dst, msg, r.Min.X, r.Min.Y)
}
