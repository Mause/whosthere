package components

import (
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/rivo/tview"
)

const (
	Divider = " | "
)

type UIComponent interface {
	tview.Primitive
	Render(s state.ReadOnly)
}
