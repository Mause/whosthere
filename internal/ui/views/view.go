package views

import (
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/rivo/tview"
)

type View interface {
	components.UIComponent
	FocusTarget() tview.Primitive
}
