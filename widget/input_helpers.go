package widget

import (
	"math"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/bstkhq/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
)

// blink ticks returns the length of "blink on", which equals "blink off"
func blinkTicks(theme *uikit.Theme) int {
	return int(math.Max(1, float64(theme.CaretBlink*time.Duration(ebiten.TPS()))/float64(time.Second)))
}

// removeLastRune removes the last UTF-8 rune from the provided string.
func removeLastRune(s string) string {
	if s == "" {
		return ""
	}
	_, sz := utf8.DecodeLastRuneInString(s)
	if sz <= 0 || sz > len(s) {
		return ""
	}
	return s[:len(s)-sz]
}

type runeClass int

const (
	runeNone runeClass = iota
	runeSpace
	runeAlphanum
	runeConnector
	runePunctuation
)

func runeBreakClass(r rune) runeClass {
	for _, class := range []runeClass{runeSpace, runeAlphanum, runeConnector, runePunctuation} {
		if isRuneClass(r, class) {
			return class
		}
	}
	return runeNone
}

func isRuneClass(r rune, class runeClass) bool {
	switch class {
	case runeSpace:
		return unicode.In(r, unicode.Space)
	case runeAlphanum:
		return unicode.In(r, unicode.Letter, unicode.Number)
	case runeConnector:
		return unicode.In(r, unicode.Pc)
	case runePunctuation:
		return unicode.In(r, unicode.P) && !unicode.In(r, unicode.Pc)
	case runeNone:
		return !unicode.In(r, unicode.Letter, unicode.Number, unicode.P)
	default:
		panic(class)
	}
}

func prevBreakPos(runes []rune, pos int) int {
	if pos <= 1 || pos > len(runes) {
		return 0
	}

	pos -= 1
	if unicode.IsSpace(runes[pos]) || isRuneClass(runes[pos], runeConnector) {
		pos -= 1
		if pos <= 1 {
			return 0
		}
	}

	matchClass := runeBreakClass(runes[pos])
	for pos > 0 {
		if isRuneClass(runes[pos-1], matchClass) {
			pos -= 1
			continue
		}
		return pos
	}
	return 0
}

func nextBreakPos(runes []rune, pos int) int {
	if pos >= len(runes) {
		return len(runes)
	}

	if unicode.IsSpace(runes[pos]) || isRuneClass(runes[pos], runeConnector) {
		pos += 1
		if pos >= len(runes) {
			return len(runes)
		}
	}

	matchClass := runeBreakClass(runes[pos])
	for pos < len(runes) {
		if isRuneClass(runes[pos], matchClass) {
			pos += 1
			continue
		}
		return pos
	}
	return len(runes)
}
