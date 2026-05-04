package mobile

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"

	"github.com/bstkhq/go-uikit/demo"
	"github.com/hajimehoshi/ebiten/v2/mobile"
)

var g = demo.New()
var once sync.Once

func SetAndroidID(id int) {
	once.Do(func() { go run() })
}

func run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[PANIC] %s: %s\n", r, debug.Stack())
			os.Exit(1)
		}
	}()

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
