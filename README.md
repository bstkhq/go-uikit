# go-uikit

A small immediate-mode UI kit for [Ebiten](https://ebitengine.org/) games and apps: labels, buttons,
checkboxes, text inputs, a text area, a dropdown select, and layout containers (stack, row-stack, grid),
all sharing one rule — **every proportion is derived from the font**, so nothing needs to be tuned by hand.

**[▶ Try the live demo](https://bstkhq.github.io/go-uikit/)** — the exact
`cmd/demo` binary below, compiled to WebAssembly and running in your browser.

## Contents

- [Design philosophy](#design-philosophy)
- [Installation](#installation)
- [Quick start](#quick-start)
- [Core concepts](#core-concepts)
- [Widgets](#widgets)
- [Layouts](#layouts)
- [Events](#events)
- [Validation](#validation)
- [Scrolling](#scrolling)
- [Focus & keyboard navigation](#focus--keyboard-navigation)
- [Text input & IME (soft keyboards)](#text-input--ime-soft-keyboards)
- [Building for other platforms](#building-for-other-platforms)
- [Examples in this repo](#examples-in-this-repo)
- [License](#license)

## Design philosophy

- All widgets share the same control height, derived from font metrics.
- External layout can only control X/Y and width. Height is always computed by the theme (plus
  an optional error line for invalid fields).
- No magic numbers: padding, radius, border width, gaps, etc. are all derived from the control height,
  not set individually per widget.

The result is that swapping `uikit.NewTheme(font, fontPx)` for a different font or size rescales the
whole UI consistently, instead of a pile of independently-tuned constants drifting apart.

## Installation

```bash
go get github.com/bstkhq/go-uikit
```

Requires Go 1.24+ and [Ebiten v2](https://github.com/hajimehoshi/ebiten).

## Quick start

```go
package main

import (
	"log"

	"github.com/bstkhq/go-uikit"
	"github.com/bstkhq/go-uikit/layout"
	"github.com/bstkhq/go-uikit/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	ctx *uikit.Context
}

func NewGame() *Game {
	theme := uikit.DefaultTheme()

	root := layout.NewStack(theme)
	root.SetPadding(theme.SpaceM, theme.SpaceM)

	label := widget.NewLabel(theme, "Hello, go-uikit!")

	button := widget.NewButton(theme, "Click me")
	button.On(uikit.EventClick, func(uikit.Event) bool {
		label.SetText("Clicked!")
		return false
	}, false)

	root.Add(label, button)

	return &Game{ctx: uikit.NewContext(theme, root, nil)}
}

func (g *Game) Update() error                       { g.ctx.Update(); return nil }
func (g *Game) Draw(screen *ebiten.Image)            { g.ctx.Draw(screen) }
func (g *Game) Layout(w, h int) (int, int)           { return w, h }

func main() {
	ebiten.SetWindowSize(480, 320)
	ebiten.SetWindowTitle("go-uikit quick start")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
```

## Core concepts

| Type | Role |
|---|---|
| `uikit.Theme` | The single source of truth for proportions and colors. Built from a font + pixel size via `uikit.NewTheme(font, fontPx)`, or `uikit.DefaultTheme()` for a ready-made dark theme (Go's built-in `goregular` at 20px). All fields are exported, so you can mutate the returned `*Theme` to restyle colors without touching layout math. |
| `uikit.Context` | Owns the widget tree, routes pointer/keyboard input, tracks focus, and drives per-frame `Update`/`Draw`. One per `ebiten.Game`. Created with `uikit.NewContext(theme, root, ime)`. |
| `uikit.Widget` | The interface every control implements: frame placement, hover/press/focus/enabled/visible state, `Measure`, `Update`, `Draw`, and event registration. |
| `uikit.Layout` | A `Widget` that owns children (`Add`, `SetChildren`, `Clear`) and can draw overlays above them (used by `Select`'s dropdown). Stack, RowStack and Grid all implement it — and since a `Layout` is itself a `Widget`, layouts nest freely. |
| `uikit.Base` | Embedded by every widget/layout. Supplies the shared hover/press/focus/enabled/invalid bookkeeping, the standard surface/border/focus-ring/error drawing, and height computation from the theme. |
| `uikit.EventDispatcher` | Embedded via `Base`. Backs `On`/`Dispatch` for every widget. |

## Widgets

All constructors take `*uikit.Theme` first. All widgets embed `uikit.Base` and satisfy `uikit.Widget`.

| Widget | Constructor | Notes |
|---|---|---|
| `widget.Label` | `NewLabel(theme, text)` | Wraps text to its width; `SetText`, or `SetTextFunc(func() string)` for a value that's recomputed every frame (handy for live stats). `SetTextModifiers(...)` applies one-off `widget.Font`/`widget.Color`/`widget.Size` overrides to the renderer. |
| `widget.Button` | `NewButton(theme, label)` | Fires `EventClick` on pointer click, and also on Enter/Space while focused. `SetLabel` to change text. |
| `widget.Checkbox` | `NewCheckbox(theme, label)` | `SetChecked`/`Checked()`; fires `EventValueChange`. Toggles on click or Space while focused. |
| `widget.TextInput` | `NewTextInput(theme, placeholder)` | Single-line text with caret, selection-free editing, and blinking cursor. `Text()`/`SetText()`, `SetInputRuneLimit(n)`, `Caret()`/`SetCaret(i)`. Set `.IMEOptions` to hint the on-screen keyboard type (see [IME](#text-input--ime-soft-keyboards)). Fires `EventValueChange`. |
| `widget.TextArea` | `NewTextArea(theme, placeholder)` | Multi-line variant; `SetLines(n)` sets the visible height in lines. |
| `widget.Select` | `NewSelect(theme, []widget.SelectOption)` | Dropdown with `{Value any; Label string}` options. `SetOptions`, `Index()`/`SetIndex(i)`, `Value()`, `Selected() (SelectOption, bool)`, `SetPlaceholder`, and `MaxVisible` (rows shown when open). The open list is drawn as an overlay, so it never shifts sibling widgets. |
| `widget.Container` | `NewContainer(theme)` | An empty themed box for custom content. `SetHeight(px)`, then hook `OnUpdate(ctx, contentRect)` / `OnDraw(ctx, dst)` — `dst` is already a `SubImage` clipped to the padded content area. |

## Layouts

All layouts live in the `layout` package, implement `uikit.Layout`, and can be added to a `Context` or
nested inside one another as regular widgets.

| Layout | Constructor | Notes |
|---|---|---|
| `layout.Stack` | `NewStack(theme)` | Vertical list. `SetPadding(x, y)`, `SetGap(v)`. `SetHeight(0)` (default) sizes to content; `SetHeight(px)` fixes a viewport and enables scrolling/clipping. |
| `layout.RowStack` | `NewRowStack(theme, weights...)` | Rows of independently-weighted columns — e.g. `NewRowStack(theme, 2, 1)` gives every row a 2:1 split by default. Override a single row with `SetRowPattern(rowIndex, weights...)`. `SetGap(h, v)`, `SetAlign(layout.AlignStart/AlignMiddle/AlignEnd)` for vertical alignment within a row, and a `BeforeUpdate` hook to adjust children just before layout runs. |
| `layout.Grid` | `NewGrid(theme)` | Fixed column count (`SetColumns(n)`, default 2), all cells equal width. `SetGap(x, y)`, `SetPadding(x, y)`, `SetHeight(px)` for scrolling. |

## Events

```go
type EventType int
const (
	EventFocusGained
	EventFocusLost
	EventPointerDown
	EventPointerUp
	EventClick
	EventKeyDown
	EventKeyUp
	EventValueChange
)
```

Register handlers with `widget.On(eventType, func(uikit.Event) bool { ... }, clear bool)`. Returning
`true` stops the event from reaching any other handlers registered for that type on that widget;
`clear` wipes previously-registered handlers of that type before adding the new one (useful when a
handler closure needs to be replaced rather than appended).

## Validation

Any widget can be marked invalid: `widget.SetInvalid("message")` / `widget.ClearInvalid()` /
`widget.IsInvalid() (bool, string)`. An invalid widget gets an error-colored border and an error line
drawn below it, and its height grows to make room automatically — layouts don't need special-casing.
See `g.txtB` and `g.sel` in `demo/game.go` for a required-field example wired to `EventValueChange`.

## Scrolling

`Stack`, `RowStack` and `Grid` all embed `uikit.Scroller`. Give any of them a fixed height via
`SetHeight(px)` and they become scrollable: mouse wheel, and press-drag on both desktop and touch,
with a themed scrollbar that (by default, `uikit.ScrollbarOnMove`) only appears while scrolling.

## Focus & keyboard navigation

`Context.Update()` handles Tab/Shift+Tab to cycle focus across visible, enabled, focusable widgets,
and routes pointer clicks to focus the topmost hit widget. Widgets opt in via `Focusable() bool`
(layouts return `false`; interactive controls return `true`).

## Text input & IME (soft keyboards)

`TextInput` and `TextArea` implement `uikit.IME`, exposing an `IMEOptions` bitmask you set to hint the
platform keyboard: keyboard kind (`KeyboardText`, `KeyboardNumber`, `KeyboardEmail`, `KeyboardPhone`,
`KeyboardURI`, `KeyboardPassword`, `KeyboardMultiline`), the return-key action
(`ActionGo`/`ActionSearch`/`ActionSend`/`ActionNext`/`ActionDone`), capitalization
(`CapsSentences`/`CapsWords`/`CapsAll`), and flags like `NoSuggestions`/`NoPersonalizedLearning`.

On desktop these are mostly no-ops (real keyboard input works regardless); on mobile, wire a platform
`uikit.IMEBridge` via `Context.SetIMEBridge` — see `cmd/android` for the Ebiten Mobile integration that
shows/hides the Android keyboard and syncs IME composing text.

## Building for other platforms

**Desktop:**
```bash
go run ./cmd/demo
```

**WebAssembly** — this is exactly how the [live demo](https://bstkhq.github.io/go-uikit/) above was built,
and is served straight from [`docs/`](docs) via GitHub Pages:
```bash
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -trimpath -o docs/demo.wasm ./cmd/demo
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" docs/   # older Go: misc/wasm/wasm_exec.js
touch docs/.nojekyll                                # skip Jekyll processing on Pages
```
`docs/index.html` loads the glue script and does
`WebAssembly.instantiateStreaming(fetch("demo.wasm"), go.importObject).then(r => go.run(r.instance))`,
with a plain fallback for servers/browsers that don't stream-compile. Ebiten creates and manages its
own `<canvas>` — no other markup is required. To publish: push `docs/` to `master` and enable
**Settings → Pages → Deploy from a branch → master /docs** on GitHub.

**Android:** see [`cmd/android/README.md`](cmd/android/README.md) for the Ebiten Mobile +
`apk-ebiten-builder` toolchain, including the IME bridge wiring.

## Examples in this repo

| Path | What it shows |
|---|---|
| [`cmd/demo`](cmd/demo) (source in [`demo/game.go`](demo/game.go)) | Every widget and both `Stack`/`Grid` layouts together, with live focus/TPS/FPS labels, validation, and a custom `Container`. This is the binary behind the live WASM demo. |
| [`cmd/rowstack`](cmd/rowstack) | `RowStack` layout patterns — mixed column weights per row. |
| [`cmd/android`](cmd/android) / [`example/android`](example/android) | Packaging the demo as an Android APK via Ebiten Mobile, with the IME bridge connected to the system keyboard. |

## License

MIT — see [LICENSE](LICENSE).
