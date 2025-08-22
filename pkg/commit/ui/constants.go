package ui

// Checkbox IDs
const (
	CheckboxSign    = "sign"
	CheckboxPush    = "push"
	CheckboxHooks   = "tag_major"
	CheckboxVerbose = "tag_minor"
	CheckboxAmend   = "tag_patch"
)

var checkboxDefaults = map[string]bool{
	CheckboxSign:    false,
	CheckboxPush:    false,
	CheckboxHooks:   false,
	CheckboxVerbose: false,
	CheckboxAmend:   false,
}

// UI Text
const (
	ListTitle         = "Select Commit Message"
	ManualOptionTitle = "Write custom message"
	ManualOptionDesc  = "Enter your own commit message"
	ManualInputTitle  = "Write Your Commit Message"
	ManualInputHelp   = "Enter: new line • Ctrl+D: finish • Esc: cancel"
	FooterHelp        = "Press 1-5 to toggle options"
	ProviderManual    = "manual"
)

// Checkbox Labels
const (
	LabelSign    = "Sign commit"
	LabelPush    = "Push to remote"
	LabelHooks   = "Tag: major"
	LabelVerbose = "Tag: minor"
	LabelAmend   = "Tag: patch"
)

// Unicode Characters
const (
	CheckboxChecked   = "▣"
	CheckboxUnchecked = "▢"
	Cursor            = "│"
)

// Colors (lipgloss color codes)
const (
	ColorPrimary      = "170" // Purple
	ColorSecondary    = "245" // Light gray
	ColorNormal       = "250" // White-ish
	ColorDimmed       = "240" // Gray
	ColorDimmedDark   = "238" // Dark gray
	ColorDimmedDarker = "236" // Darker gray
	ColorBorder       = "240" // Border gray
	ColorAccent       = "62"  // Accent color
	ColorBright       = "230" // Bright
	ColorMuted        = "241" // Muted gray
)

// Layout Constants
const (
	PaddingTop         = 2
	PaddingHorizontal  = 4
	FooterHeightApprox = 10 // Approximate height needed for footer
	DefaultListHeight  = 10
	MaxListHeight      = 15
	MinListHeight      = 3
	MaxDisplayLines    = 10 // Max lines to show in multi-line preview
	MaxDescriptionLen  = 60 // Max length for single-line description
	ManualInputWidth   = 80
	ManualInputHeight  = 10
)

// Keybindings
const (
	KeyQuit        = "q"
	KeySelect      = "enter"
	KeyCancel      = "esc"
	KeyNewLine     = "enter"
	KeyFinishInput = "ctrl+d"
	KeyBackspace   = "backspace"
	KeySpace       = " "
	KeyInterrupt   = "ctrl+c"
)

// Checkbox toggle keys
const (
	Key1 = "1"
	Key2 = "2"
	Key3 = "3"
	Key4 = "4"
	Key5 = "5"
)
