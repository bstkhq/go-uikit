package uikit

// IMEBridge is implemented on the Java side and registered from your mobile package.
// In this minimal integration, it only opens/closes the keyboard.
type IMEBridge interface {
	Show(opts IMEOptions)
	Hide()
}

// IME marks widgets that should trigger IME when focused.
type IME interface {
	IME() IMEOptions
}

// IMEOptions control the IME input type and options. There are three
// parts that can be controlled: the keyboard type, the submit icon and the
// capitalization. Options can be combined with bitwise ORs:
//
//	opts := KeyboardEmail // keyboard includes @ and .com more prominently
//	opts := KeyboardURI | ActionSearch // combine two options
//	opts := KeyboardText | CapSentences // add capitalization
//	opts := KeyboardText | ActionSend | CapsSentences // all options
type IMEOptions = int32

const (
	// Keyboard Types
	KeyboardText      IMEOptions = 0x000
	KeyboardMultiline IMEOptions = 0x100
	KeyboardNumber    IMEOptions = 0x200
	KeyboardPhone     IMEOptions = 0x300
	KeyboardEmail     IMEOptions = 0x400
	KeyboardURI       IMEOptions = 0x500
	KeyboardPassword  IMEOptions = 0x600

	// action type (defaults to return), typically affects the icon
	ActionGo     IMEOptions = 0x010
	ActionSearch IMEOptions = 0x020
	ActionSend   IMEOptions = 0x030
	ActionNext   IMEOptions = 0x040
	ActionDone   IMEOptions = 0x050

	// capitalization (defaults to none)
	CapsSentences IMEOptions = 0x001
	CapsWords     IMEOptions = 0x002
	CapsAll       IMEOptions = 0x003
)
