package uikit

// IMEBridge is implemented on the Java side and registered from your mobile package.
// In this minimal integration, it only opens/closes the keyboard.
type IMEBridge interface {
	// Android values can be derived with [IMEOptions.AndroidParameters]().
	Show(inputType, imeOptions int32)
	Hide()
}

// IME marks widgets that should trigger IME when focused.
type IME interface {
	IME() IMEOptions
}

// IMEOptions control the IME input type and options. There are three
// parts that can be controlled: the keyboard type, the submit/action icon
// and capitalization. Options can be combined with bitwise ORs:
//
//	opts := KeyboardEmail // keyboard includes @ and .com more prominently
//	opts := KeyboardURI | ActionSearch // combine two options
//	opts := KeyboardText | CapSentences // add capitalization
//	opts := KeyboardText | ActionSend | CapsSentences // all options
//
// Notice that some keyboards can't support some flags and might be downgraded
// at runtime. For example, samsung IME won't support KeyboardRaw.
type IMEOptions int32

// IMEOptions constants.
const (
	// keyboard Types
	KeyboardRaw       IMEOptions = 0x000 // no configuration, most compatible with key events
	KeyboardText      IMEOptions = 0x100
	KeyboardMultiline IMEOptions = 0x200
	KeyboardNumber    IMEOptions = 0x300
	KeyboardPhone     IMEOptions = 0x400
	KeyboardEmail     IMEOptions = 0x500
	KeyboardURI       IMEOptions = 0x600
	KeyboardPassword  IMEOptions = 0x700

	// action type, typically affects the icon (defaults to return key)
	ActionGo     IMEOptions = 0x010
	ActionSearch IMEOptions = 0x020
	ActionSend   IMEOptions = 0x030
	ActionNext   IMEOptions = 0x040
	ActionDone   IMEOptions = 0x050

	// capitalization (defaults to none)
	CapsSentences IMEOptions = 0x001
	CapsWords     IMEOptions = 0x002
	CapsAll       IMEOptions = 0x003

	// other flags

	// NoSuggestions is intended only for KeyboardText,
	// and is ignored by most IMEs in practice.
	NoSuggestions IMEOptions = 0x1000

	// NoPersonalizedLearning enters an "incognito" mode, often
	// including icons or a color change to reflect it.
	NoPersonalizedLearning IMEOptions = 0x2000
)

const (
	android_TYPE_MASK_CLASS int32 = 0x0000000F

	// classes
	android_TYPE_CLASS_TEXT     int32 = 0x00000001
	android_TYPE_CLASS_NUMBER   int32 = 0x00000002
	android_TYPE_CLASS_PHONE    int32 = 0x00000003
	android_TYPE_CLASS_DATETIME int32 = 0x00000004

	// flags for TYPE_CLASS_TEXT
	android_TYPE_TEXT_FLAG_MULTI_LINE     int32 = 0x00020000
	android_TYPE_TEXT_FLAG_CAP_SENTENCES  int32 = 0x00004000
	android_TYPE_TEXT_FLAG_CAP_WORDS      int32 = 0x00002000
	android_TYPE_TEXT_FLAG_CAP_CHARACTERS int32 = 0x00001000
	android_TYPE_TEXT_FLAG_NO_SUGGESTIONS int32 = 0x00080000

	// variations for TYPE_CLASS_TEXT
	android_TYPE_TEXT_VARIATION_EMAIL_ADDRESS int32 = 0x00000020
	android_TYPE_TEXT_VARIATION_URI           int32 = 0x00000010
	android_TYPE_TEXT_VARIATION_PASSWORD      int32 = 0x00000080

	// IME Actions (EditorInfo.imeOptions)
	android_IME_ACTION_GO     int32 = 0x00000002
	android_IME_ACTION_SEARCH int32 = 0x00000003
	android_IME_ACTION_SEND   int32 = 0x00000004
	android_IME_ACTION_NEXT   int32 = 0x00000005
	android_IME_ACTION_DONE   int32 = 0x00000006

	// misc
	andoid_IME_FLAG_NO_PERSONALIZED_LEARNING int32 = 0x01000000
)

// AndroidParameters returns the inputType and imeOptions to be
// sent to IMEBridge.Show().
func (o IMEOptions) AndroidParameters() (int32, int32) {
	inputType := o.extractInputType()
	inputType = o.applyCapitalization(inputType)
	imeOptions := o.extractImeOptions()

	if o&NoSuggestions == NoSuggestions {
		inputType |= android_TYPE_TEXT_FLAG_NO_SUGGESTIONS
	}
	if o&NoPersonalizedLearning == NoPersonalizedLearning {
		imeOptions |= andoid_IME_FLAG_NO_PERSONALIZED_LEARNING
	}
	return inputType, imeOptions
}

func (o IMEOptions) extractInputType() int32 {
	switch o & 0xF00 {
	case 0x100:
		return android_TYPE_CLASS_TEXT
	case 0x200:
		return android_TYPE_CLASS_TEXT | android_TYPE_TEXT_FLAG_MULTI_LINE
	case 0x300:
		return android_TYPE_CLASS_NUMBER
	case 0x400:
		return android_TYPE_CLASS_PHONE
	case 0x500:
		return android_TYPE_CLASS_TEXT | android_TYPE_TEXT_VARIATION_EMAIL_ADDRESS
	case 0x600:
		return android_TYPE_CLASS_TEXT | android_TYPE_TEXT_VARIATION_URI
	case 0x700:
		return android_TYPE_CLASS_TEXT | android_TYPE_TEXT_VARIATION_PASSWORD | android_TYPE_TEXT_FLAG_NO_SUGGESTIONS
	default:
		return 0
	}
}

func (o IMEOptions) applyCapitalization(inputType int32) int32 {
	if inputType&android_TYPE_MASK_CLASS != android_TYPE_CLASS_TEXT || inputType&android_TYPE_TEXT_VARIATION_PASSWORD == android_TYPE_TEXT_VARIATION_PASSWORD {
		return inputType
	}

	switch o & 0x00F {
	case 0x001:
		return inputType | android_TYPE_TEXT_FLAG_CAP_SENTENCES
	case 0x002:
		return inputType | android_TYPE_TEXT_FLAG_CAP_WORDS
	case 0x003:
		return inputType | android_TYPE_TEXT_FLAG_CAP_CHARACTERS
	default:
		return inputType
	}
}

func (o IMEOptions) extractImeOptions() int32 {
	switch o & 0x0F0 {
	case 0x010:
		return android_IME_ACTION_GO
	case 0x020:
		return android_IME_ACTION_SEARCH
	case 0x030:
		return android_IME_ACTION_SEND
	case 0x040:
		return android_IME_ACTION_NEXT
	case 0x050:
		return android_IME_ACTION_DONE
	default:
		return 0
	}
}
