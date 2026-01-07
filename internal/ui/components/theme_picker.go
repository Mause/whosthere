package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

// ThemePicker is a modal for selecting and previewing themes.
type ThemePicker struct {
	*tview.Modal
	list          *tview.List
	footer        *tview.TextView
	themes        []string
	currentIndex  int
	originalTheme string
	onSelect      func(themeName string)
	onSave        func(themeName string)
	onCancel      func()
	themeManager  *theme.Manager
}

// NewThemePicker creates a new theme picker modal.
func NewThemePicker(tm *theme.Manager) *ThemePicker {
	footer := tview.NewTextView()
	footer.SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("j/k: navigate | Enter: apply | Shift+Enter: save to config | Esc: cancel")
	footer.SetTextColor(tview.Styles.SecondaryTextColor)
	footer.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	theme.RegisterPrimitive(footer) // Register footer with theme manager

	tp := &ThemePicker{
		Modal:        tview.NewModal(),
		list:         tview.NewList(),
		footer:       footer,
		themes:       theme.Names(),
		themeManager: tm,
	}

	tp.buildList()
	theme.RegisterPrimitive(tp.list) // Register list with theme manager
	tp.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	return tp
}

// buildList populates the list with available themes.
func (tp *ThemePicker) buildList() {
	tp.list.Clear()
	tp.list.SetBorder(true).
		SetTitle(" Theme Picker - Preview themes live ").
		SetTitleAlign(tview.AlignCenter).
		SetTitleColor(tview.Styles.TitleColor).
		SetBorderColor(tview.Styles.BorderColor).
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	// Remove secondary text to reduce spacing between items
	tp.list.ShowSecondaryText(false)

	currentTheme := tp.themeManager.Current()

	for i, themeName := range tp.themes {
		displayName := themeName
		if themeName == currentTheme {
			displayName = fmt.Sprintf("● %s (current)", themeName)
			tp.currentIndex = i
		}

		// Capture the theme name for the closure
		name := themeName
		// Don't provide secondary text to keep items compact
		tp.list.AddItem(displayName, "", 0, func() {
			if tp.onSelect != nil {
				tp.onSelect(name)
			}
		})
	}

	tp.list.SetCurrentItem(tp.currentIndex)
	tp.setupInputHandling()
}

// setupInputHandling configures vim-style navigation and preview.
func (tp *ThemePicker) setupInputHandling() {
	tp.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Rune() == 'j' || event.Key() == tcell.KeyDown:
			// Move down and preview
			nextIdx := tp.list.GetCurrentItem() + 1
			if nextIdx < len(tp.themes) {
				tp.list.SetCurrentItem(nextIdx)
				tp.previewTheme(tp.themes[nextIdx])
			}
			return nil
		case event.Rune() == 'k' || event.Key() == tcell.KeyUp:
			// Move up and preview
			prevIdx := tp.list.GetCurrentItem() - 1
			if prevIdx >= 0 {
				tp.list.SetCurrentItem(prevIdx)
				tp.previewTheme(tp.themes[prevIdx])
			}
			return nil
		case event.Key() == tcell.KeyEnter && event.Modifiers()&tcell.ModShift != 0:
			// Shift+Enter: Save to config
			currentIdx := tp.list.GetCurrentItem()
			if currentIdx >= 0 && currentIdx < len(tp.themes) {
				if tp.onSave != nil {
					tp.onSave(tp.themes[currentIdx])
				}
			}
			return nil
		case event.Key() == tcell.KeyEnter:
			// Confirm selection
			currentIdx := tp.list.GetCurrentItem()
			if currentIdx >= 0 && currentIdx < len(tp.themes) {
				if tp.onSelect != nil {
					tp.onSelect(tp.themes[currentIdx])
				}
			}
			return nil
		case event.Key() == tcell.KeyEsc || event.Rune() == 'q':
			// Cancel and restore original theme
			if tp.themeManager != nil && tp.originalTheme != "" {
				tp.themeManager.SetTheme(tp.originalTheme)
			}
			if tp.onCancel != nil {
				tp.onCancel()
			}
			return nil
		}
		return event
	})
}

// previewTheme temporarily applies a theme for preview.
func (tp *ThemePicker) previewTheme(themeName string) {
	if tp.themeManager != nil {
		tp.themeManager.SetTheme(themeName)
		// Rebuild list to update the "current" marker
		tp.rebuildList()
	}
}

// rebuildList rebuilds the list items with updated theme marker.
func (tp *ThemePicker) rebuildList() {
	currentTheme := tp.themeManager.Current()
	currentSelection := tp.list.GetCurrentItem()

	tp.list.Clear()

	// Apply theme colors before adding items so they inherit correct colors
	theme.ApplyToPrimitive(tp.list)

	for _, themeName := range tp.themes {
		displayName := themeName
		if themeName == currentTheme {
			displayName = fmt.Sprintf("● %s (current)", themeName)
		}

		name := themeName
		tp.list.AddItem(displayName, "", 0, func() {
			if tp.onSelect != nil {
				tp.onSelect(name)
			}
		})
	}

	// Restore selection
	if currentSelection >= 0 && currentSelection < len(tp.themes) {
		tp.list.SetCurrentItem(currentSelection)
	}
}

// OnSelect registers a callback for when a theme is selected (Enter).
func (tp *ThemePicker) OnSelect(fn func(themeName string)) {
	tp.onSelect = fn
}

// OnSave registers a callback for when a theme should be saved to config (Shift+Enter).
func (tp *ThemePicker) OnSave(fn func(themeName string)) {
	tp.onSave = fn
}

// OnCancel registers a callback for when the picker is cancelled (Esc).
func (tp *ThemePicker) OnCancel(fn func()) {
	tp.onCancel = fn
}

// Show displays the theme picker and stores the current theme for potential rollback.
func (tp *ThemePicker) Show() {
	if tp.themeManager != nil {
		tp.originalTheme = tp.themeManager.Current()
	}
	tp.buildList()
}

// GetList returns a flex container with the list and footer for rendering.
func (tp *ThemePicker) GetList() tview.Primitive {
	container := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tp.list, 0, 1, true).
		AddItem(tp.footer, 1, 0, false)
	return container
}

// GetListPrimitive returns the actual list primitive for focus management.
func (tp *ThemePicker) GetListPrimitive() *tview.List {
	return tp.list
}

// GetFooter returns the footer primitive.
func (tp *ThemePicker) GetFooter() *tview.TextView {
	return tp.footer
}
