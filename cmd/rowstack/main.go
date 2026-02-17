package main

import (
	"log"
	"strconv"

	"github.com/erparts/go-uikit"
	"github.com/erparts/go-uikit/layout"
	"github.com/erparts/go-uikit/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowTitle("uikit - RowStack")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	var g Game
	g.Initialize()
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	ctx *uikit.Context
	ime uikit.IMEBridge
}

func (g *Game) SetIMEBridge(b uikit.IMEBridge) {
	g.ime = b
	if g.ctx != nil {
		g.ctx.SetIMEBridge(b)
	}
}

func (g *Game) Initialize() {
	theme := uikit.DefaultTheme()
	stack := layout.NewStack(theme)
	stack.SetPadding(theme.SpaceS, theme.SpaceS)
	g.ctx = uikit.NewContext(theme, stack, g.ime)

	patternsOpts := []widget.SelectOption{
		{Label: "1", Value: []float64{1.0}},
		{Label: "1 | 1", Value: []float64{1.0, 1.0}},
		{Label: "1 | 2 | 1", Value: []float64{1.0, 2.0, 1.0}},
		{Label: "3 | 2 | 1", Value: []float64{3.0, 2.0, 1.0}},
		{Label: "1 | 1.5 | 1.5 | 1.0", Value: []float64{1.0, 1.5, 1.5, 1.0}},
	}

	stack.Add(widget.NewLabel(theme, "Row stack configuration:"))
	configRows := layout.NewRowStack(theme, 1.0, 2.0)
	configRows.SetAlign(layout.AlignMiddle)
	selectPatternDefault := widget.NewSelect(theme, patternsOpts)
	selectPatternRow1 := widget.NewSelect(theme, patternsOpts)
	selectPatternRow3 := widget.NewSelect(theme, patternsOpts)
	configRows.Add(widget.NewLabel(theme, "Default Pattern"), selectPatternDefault)
	configRows.Add(widget.NewLabel(theme, "First Row Pattern"), selectPatternRow1)
	configRows.Add(widget.NewLabel(theme, "Third Row Pattern"), selectPatternRow3)
	stack.Add(configRows)

	stack.Add(widget.NewLabel(theme, "\n"))
	stack.Add(widget.NewLabel(theme, "Row stack:"))
	rowStack := layout.NewRowStack(theme)
	for i := range 16 {
		rowStack.Add(widget.NewButton(theme, "Button #"+strconv.Itoa(i)))
	}
	stack.Add(rowStack)

	onPatternChange := func(ev uikit.Event, targetRow int) bool {
		pattern := ev.Widget.(*widget.Select).Value().([]float64)
		rowStack.SetRowPattern(targetRow, pattern...)
		return true
	}
	selectPatternDefault.On(uikit.EventValueChange, uikit.EventHandler(func(ev uikit.Event) bool {
		return onPatternChange(ev, -1)
	}), false)
	selectPatternRow1.On(uikit.EventValueChange, uikit.EventHandler(func(ev uikit.Event) bool {
		return onPatternChange(ev, 0)
	}), false)
	selectPatternRow3.On(uikit.EventValueChange, uikit.EventHandler(func(ev uikit.Event) bool {
		return onPatternChange(ev, 2)
	}), false)
}

func (g *Game) Layout(w, h int) (int, int) {
	return w, h
}

func (g *Game) Update() error {
	g.ctx.Update()
	return nil
}

func (g *Game) Draw(canvas *ebiten.Image) {
	g.ctx.Draw(canvas)
}
