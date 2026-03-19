package demo

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"

	"github.com/bstkhq/go-uikit"
	"github.com/bstkhq/go-uikit/layout"
	"github.com/bstkhq/go-uikit/widget"
)

type Game struct {
	stack *layout.Stack
	grid  *layout.Grid
	ime   uikit.IMEBridge

	theme *uikit.Theme
	ctx   *uikit.Context

	title        *widget.Label
	txtA         *widget.TextInput
	numA         *widget.TextInput
	txtB         *widget.TextInput
	txtDis       *widget.TextInput
	ta           *widget.TextArea
	sel          *widget.Select
	box          *widget.Container
	chkA         *widget.Checkbox
	chkDis       *widget.Checkbox
	chkGrid      *widget.Checkbox
	btnA         *widget.Button
	btnDis       *widget.Button
	focusInfo    *widget.Label
	exampleLabel *widget.Label

	clickCount int
}

func New() *Game {
	return &Game{}
}

// SetIMEBridge can be called from mobile bindings to enable keyboard show/hide.
func (g *Game) SetIMEBridge(b uikit.IMEBridge) {
	g.ime = b
	if g.ctx != nil {
		g.ctx.SetIMEBridge(b)
	}
}

func (g *Game) initOnce() {
	if g.ctx != nil {
		return
	}

	g.theme = uikit.DefaultTheme()

	root := layout.NewStack(g.theme)
	root.SetPadding(g.theme.SpaceS, g.theme.SpaceS)
	g.ctx = uikit.NewContext(g.theme, root, g.ime)
	g.stack = layout.NewStack(g.theme)

	g.grid = layout.NewGrid(g.theme)
	g.grid.SetVisible(false)

	g.title = widget.NewLabel(g.theme, "")
	g.title.SetTextFunc(func() string {
		return fmt.Sprintf("UI Kit Demo (TPS: %0.2f - FPS: %0.2f)", ebiten.ActualTPS(), ebiten.ActualFPS())
	})

	g.focusInfo = widget.NewLabel(g.theme, "")
	g.focusInfo.SetTextFunc(func() string {
		if fw := g.ctx.Focused(); fw != nil {
			return fmt.Sprintf("Focused: %T", fw)
		}

		return "Focused: (none) — tap a widget or TAB"
	})

	g.exampleLabel = widget.NewLabel(g.theme, "Label example: static helper text")

	g.txtA = widget.NewTextInput(g.theme, "Type here…")
	g.numA = widget.NewTextInput(g.theme, "Numeric input...")
	g.numA.IMEOptions = uikit.KeyboardNumber

	g.txtB = widget.NewTextInput(g.theme, "Search…")
	g.txtB.On(uikit.EventValueChange, func(e uikit.Event) bool {
		v := e.Widget.(*widget.TextInput).Text()
		if v == "" {
			g.txtB.SetInvalid("Required")
			return false
		}

		g.txtB.ClearInvalid()
		return false
	}, false)

	g.txtDis = widget.NewTextInput(g.theme, "Disabled input")
	g.txtDis.SetText("Disabled value")
	g.txtDis.SetEnabled(false)

	g.ta = widget.NewTextArea(g.theme, "Multi-line…")
	g.ta.SetLines(5)
	g.ta.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7")

	g.sel = widget.NewSelect(g.theme, nil)
	g.sel.SetOptions([]widget.SelectOption{
		{0, "Select a value..."},
		{1, "Option A"}, {2, "Option B"}, {3, "Option C"},
		{4, "Option D"}, {5, "Option E"}, {6, "Option F"},
	})

	g.sel.On(uikit.EventValueChange, func(e uikit.Event) bool {
		s, isSelected := g.sel.Selected()
		if s.Value == 0 || !isSelected {
			g.sel.SetInvalid("Requried value")
			return false
		}

		g.sel.ClearInvalid()
		return false
	}, false)

	g.box = widget.NewContainer(g.theme)
	g.box.SetHeight(140)
	g.box.OnDraw = func(ctx *uikit.Context, dst *ebiten.Image) {
		s, _ := g.sel.Selected()
		lines := []string{
			"Custom container (user content)",
			fmt.Sprintf("- Click Count: %d", g.clickCount),
			fmt.Sprintf("- Select Value: %s ", s.Label),
			fmt.Sprintf("- Search Text: %s", g.txtB.Text()),
			fmt.Sprintf("- TextArea Chars: %d", len([]rune(g.ta.Text()))),
		}

		t := ctx.Theme().Text()
		t.SetColor(ctx.Theme().MutedTextColor)
		t.SetAlign(etxt.Left | etxt.Top)
		t.Draw(dst, strings.Join(lines, "\n"), dst.Bounds().Min.X, dst.Bounds().Min.Y)
	}

	g.chkA = widget.NewCheckbox(g.theme, "Enable main button")
	g.chkA.SetChecked(true)
	g.chkA.On(uikit.EventValueChange, func(e uikit.Event) bool {
		g.btnA.SetEnabled(e.Widget.(*widget.Checkbox).Checked())
		return false
	}, false)

	g.chkDis = widget.NewCheckbox(g.theme, "Disabled checkbox")
	g.chkDis.SetChecked(true)
	g.chkDis.SetEnabled(false)

	g.chkGrid = widget.NewCheckbox(g.theme, "Use grid layout")
	g.chkGrid.On(uikit.EventValueChange, func(e uikit.Event) bool {
		if e.Widget.(*widget.Checkbox).Checked() {
			g.grid.SetVisible(true)
			g.stack.SetVisible(false)
			return false
		}

		g.stack.SetVisible(true)
		g.grid.SetVisible(false)
		return false
	}, false)

	g.btnA = widget.NewButton(g.theme, "Action (enabled)")
	g.btnA.On(uikit.EventClick, func(_ uikit.Event) bool {
		g.clickCount++
		return false
	}, false)

	g.btnDis = widget.NewButton(g.theme, "Action (disabled)")
	g.btnDis.SetEnabled(false)

	g.ctx.Add(g.title)
	g.ctx.Add(g.focusInfo)
	g.ctx.Add(g.chkGrid)

	g.ctx.Add(g.stack)
	g.ctx.Add(g.grid)

	contentWidgets := []uikit.Widget{
		g.exampleLabel,
		g.txtA,
		g.numA,
		g.txtB,
		g.txtDis,
		g.ta,
		g.sel,
		g.box,
		g.chkA,
		g.chkDis,
		g.btnA,
		g.btnDis,
	}

	g.stack.SetChildren(contentWidgets)
	g.grid.SetChildren(contentWidgets)
}

func (g *Game) Update() error {
	g.ctx.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.ctx.Draw(screen)
}

func (g *Game) Layout(outW, outH int) (int, int) {
	g.initOnce()

	m := ebiten.Monitor()
	if m == nil {
		return 0, 0
	}

	dev := m.DeviceScaleFactor()
	if dev <= 0 {
		dev = 1
	}

	// Optional: make UI a bit larger on small screens.
	minSide := float64(outW)
	if float64(outH) < minSide {
		minSide = float64(outH)
	}

	return outW, outH
}
