package views

import (
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ View = &ThemeModalView{}

// ThemeModalView is a modal overlay page for selecting themes.
// It uses a centered flex layout to create a modal-like appearance.
type ThemeModalView struct {
	*tview.Flex
	picker *components.ThemePicker
	footer *tview.TextView

	emit func(events.Event)
}

// NewThemeModalView creates a new theme picker modal page.
// It uses the singleton theme manager, so no need to pass it in.
func NewThemeModalView(emit func(events.Event)) *ThemeModalView {
	picker := components.NewThemePicker(emit)
	footer := tview.NewTextView()
	footer.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("j/k: navigate" + components.Divider + "Enter: keep" + components.Divider + "Shift+Enter: save" + components.Divider + "Esc: cancel")
	footer.SetTextColor(tview.Styles.SecondaryTextColor)
	footer.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(picker, 0, 1, true).
		AddItem(footer, 1, 0, false)

	modalWidth := len(footer.GetText(false))

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(content, modalWidth, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	p := &ThemeModalView{
		Flex:   root,
		picker: picker,
		footer: footer,
		emit:   emit,
	}

	theme.RegisterPrimitive(content)
	theme.RegisterPrimitive(footer)

	return p
}

func (p *ThemeModalView) FocusTarget() tview.Primitive { return p.picker }

func (p *ThemeModalView) Render(s state.ReadOnly) {
	p.picker.Render(s)
}
