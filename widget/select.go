package widget

import (
	"image"
	"math"

	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

type SelectOption struct {
	Value any
	Label string
}

var _ uikit.Widget = (*Select)(nil)
var _ uikit.Hittable = (*Select)(nil)

// Select is a simple dropdown selector.
// The dropdown is rendered as an overlay (does NOT change layout of other widgets).
type Select struct {
	uikit.Base

	options []SelectOption
	index   int

	open bool

	// List scroll offset (in options, not pixels)
	scroll int

	// MaxVisible controls how many options are shown when open.
	MaxVisible int

	// optional: show when there is no selection
	placeholder string
}

func NewSelect(theme *uikit.Theme, options []SelectOption) *Select {
	cfg := uikit.NewWidgetBaseConfig(theme)

	s := &Select{
		Base:       uikit.NewBase(cfg),
		options:    options,
		index:      0,
		MaxVisible: 5,
	}

	s.clampIndex()
	s.clampScroll()

	return s
}

func (s *Select) Focusable() bool { return true }

func (s *Select) OverlayActive() bool { return s.open }

func (s *Select) SetPlaceholder(p string) { s.placeholder = p }

func (s *Select) SetOptions(opts []SelectOption) {
	s.options = opts

	// keep index in range
	s.clampIndex()
	s.clampScroll()
}

func (s *Select) Index() int { return s.index }

func (s *Select) Value() any {
	if s.index < 0 || s.index >= len(s.options) {
		return ""
	}
	return s.options[s.index].Value
}

func (s *Select) Selected() (SelectOption, bool) {
	if s.index < 0 || s.index >= len(s.options) {
		return SelectOption{}, false
	}
	return s.options[s.index], true
}

func (s *Select) SetIndex(i int) {
	if len(s.options) == 0 {
		s.index = 0
		return
	}

	if i < 0 {
		i = 0
	}
	if i >= len(s.options) {
		i = len(s.options) - 1
	}

	if s.index == i {
		return
	}

	s.index = i
	s.Dispatch(uikit.Event{Widget: s, Type: uikit.EventValueChange})

	// keep selected visible
	s.ensureIndexVisible(i)
}

func (s *Select) clampIndex() {
	if len(s.options) == 0 {
		s.index = 0
		return
	}
	if s.index < 0 {
		s.index = 0
	}
	if s.index >= len(s.options) {
		s.index = len(s.options) - 1
	}
}

func (s *Select) maxVisible() int {
	max := s.MaxVisible
	if max <= 0 {
		max = 5
	}
	return max
}

func (s *Select) maxScroll() int {
	ms := len(s.options) - s.maxVisible()
	if ms < 0 {
		return 0
	}
	return ms
}

func (s *Select) clampScroll() {
	if s.scroll < 0 {
		s.scroll = 0
	}
	ms := s.maxScroll()
	if s.scroll > ms {
		s.scroll = ms
	}
}

func (s *Select) ensureIndexVisible(idx int) {
	if idx < s.scroll {
		s.scroll = idx
	} else if idx >= s.scroll+s.maxVisible() {
		s.scroll = idx - s.maxVisible() + 1
	}
	s.clampScroll()
}

func (s *Select) listRect(ctx *uikit.Context) image.Rectangle {
	ctrl := s.Measure(false)

	n := len(s.options)
	if n <= 0 {
		n = 1
	}

	max := s.maxVisible()
	if n > max {
		n = max
	}

	listY := ctrl.Max.Y + ctx.Theme().SpaceS
	return image.Rect(ctrl.Min.X, listY, ctrl.Max.X, listY+(n*ctx.Theme().ControlH))
}

func (s *Select) HitTest(ctx *uikit.Context, pos image.Point) bool {
	ctrl := s.Measure(false)
	if pos.In(ctrl) {
		return true
	}
	if s.open {
		return pos.In(s.listRect(ctx))
	}
	return false
}

func (s *Select) Update(ctx *uikit.Context) {
	if !s.IsEnabled() {
		s.open = false
		return
	}

	ptr := ctx.Pointer()
	ctrlInside := ptr.Position.In(s.Measure(false))

	if !s.open && ptr.IsJustDown && ctrlInside {
		s.open = true
		s.ensureIndexVisible(s.index)
		return
	}

	if s.open && ptr.IsJustDown {
		list := s.listRect(ctx)

		if ptr.Position.In(list) {
			row := (ptr.Position.Y - list.Min.Y) / ctx.Theme().ControlH
			idx := s.scroll + row
			if idx >= 0 && idx < len(s.options) {
				s.SetIndex(idx)
			}
			s.open = false
			return
		}

		if ctrlInside {
			s.open = false
			return
		}

		s.open = false
		return
	}

	if s.open {
		_, wy := ebiten.Wheel()
		if wy != 0 {
			step := int(math.Copysign(1, wy))
			s.scroll -= step
			s.clampScroll()
		}
	}
}

func (s *Select) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	s.Base.Draw(ctx, dst)

	theme := ctx.Theme()
	r := s.Base.Measure(false)

	label := s.placeholder
	col := theme.MutedTextColor

	if val, ok := s.Selected(); ok {
		label = val.Label
		col = theme.TextColor
	}

	centerY := r.Min.Y + (r.Dy() / 2)

	t := theme.Text()
	t.SetColor(col)
	t.SetAlign(etxt.Left | etxt.VertCenter)
	t.Draw(dst, label, r.Min.X+theme.PadX, centerY)

	chev := "▾"
	t.SetColor(theme.TextColor)
	t.SetAlign(etxt.Right | etxt.VertCenter)
	t.Draw(dst, chev, r.Max.X-theme.PadX, centerY)
}

func (s *Select) DrawOverlay(ctx *uikit.Context, dst *ebiten.Image) {
	if !s.open {
		return
	}

	theme := ctx.Theme()
	list := s.listRect(ctx)

	s.Base.DrawRoundedRect(dst, list, theme.Radius, theme.SurfaceColor)
	s.Base.DrawRoundedBorder(dst, list, theme.Radius, theme.BorderW, theme.BorderColor)

	visibleRows := list.Dy() / theme.ControlH
	if visibleRows <= 0 {
		return
	}

	ptr := ctx.Pointer()
	for i := 0; i < visibleRows; i++ {
		idx := s.scroll + i
		if idx >= len(s.options) {
			break
		}

		y := list.Min.Y + i*theme.ControlH
		row := image.Rect(list.Min.X, y, list.Max.X, y+theme.ControlH)

		if idx == s.index {
			s.Base.DrawRoundedRect(dst, row, 0, theme.SurfaceHoverColor)
		} else if ptr.Position.In(row) {
			s.Base.DrawRoundedRect(dst, row, 0, theme.SurfaceHoverColor)
		}

		bY := row.Min.Y + row.Dy()/2
		t := theme.Text()
		t.SetColor(theme.TextColor)
		t.SetAlign(etxt.Left | etxt.VertCenter)
		t.Draw(dst, s.options[idx].Label, row.Min.X+theme.PadX, bY)
	}
}
