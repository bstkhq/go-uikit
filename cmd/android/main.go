package mobile

import (
	"github.com/bstkhq/go-uikit/demo"
	"github.com/hajimehoshi/ebiten/v2/mobile"
)

var g = demo.New()

func SetAndroidID(id int) {
	mobile.SetGame(g)
}

// IMEBridge will be generated as a Java interface that you can implement.
type IMEBridge interface {
	Show(int32, int32)
	Hide()
}

// RegisterIMEBridge is exposed to Java as Mobile.registerIMEBridge(...)
func RegisterIMEBridge(b IMEBridge) {
	g.SetIMEBridge(b)
}
