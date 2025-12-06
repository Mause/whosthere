package pages

import "github.com/derailed/tview"

// Page represents a logical screen in the TUI.
type Page interface {
	GetName() string
	GetPrimitive() tview.Primitive
	FocusTarget() tview.Primitive
}
