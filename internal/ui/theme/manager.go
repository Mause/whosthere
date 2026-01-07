package theme

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Manager coordinates theme changes using a global registry pattern.
// Primitives register themselves, and theme changes are applied to all registered primitives.
type Manager struct {
	mu         sync.RWMutex
	current    string
	app        *tview.Application
	primitives []tview.Primitive
}

var globalManager *Manager

// NewManager creates a new theme manager and sets it as the global instance.
func NewManager(app *tview.Application) *Manager {
	m := &Manager{
		app:        app,
		primitives: make([]tview.Primitive, 0),
	}
	globalManager = m
	return m
}

// Register adds a primitive to be theme-aware. When themes change, it will be updated.
func (m *Manager) Register(p tview.Primitive) {
	if p == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.primitives = append(m.primitives, p)

	// Apply current theme immediately if one is set
	if m.current != "" {
		ApplyToPrimitive(p)
	}
}

// SetTheme applies a theme by name and updates all registered primitives.
func (m *Manager) SetTheme(name string) tview.Theme {
	th := Get(name)
	tview.Styles = th

	m.mu.Lock()
	m.current = name
	primitives := append([]tview.Primitive{}, m.primitives...)
	m.mu.Unlock()

	// Apply theme to all registered primitives
	applyToAll := func() {
		for _, p := range primitives {
			ApplyToPrimitive(p)
		}
	}

	// Apply directly - tview will automatically redraw after input handler completes
	// Using QueueUpdateDraw causes deadlock when called from input handlers
	applyToAll()

	return th
}

// Current returns the currently active theme name.
func (m *Manager) Current() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// RegisterPrimitive is a convenience function to register a primitive with the global manager.
func RegisterPrimitive(p tview.Primitive) {
	if globalManager != nil {
		globalManager.Register(p)
	}
}

// ApplyToPrimitive applies theme colors to any tview primitive.
// It uses type assertions to set colors on all supported primitive types.
// This can be called manually when a primitive needs immediate theme application.
func ApplyToPrimitive(p tview.Primitive) {
	if p == nil {
		return
	}

	// Try to cast to Box (most primitives embed Box)
	if box, ok := p.(interface{ SetBackgroundColor(tcell.Color) *tview.Box }); ok {
		box.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	}

	// Try to set border color
	if bordered, ok := p.(interface{ SetBorderColor(tcell.Color) *tview.Box }); ok {
		bordered.SetBorderColor(tview.Styles.BorderColor)
	}

	// Try to set title color
	if titled, ok := p.(interface{ SetTitleColor(tcell.Color) *tview.Box }); ok {
		titled.SetTitleColor(tview.Styles.TitleColor)
	}

	// Special handling for List to set item colors
	if list, ok := p.(*tview.List); ok {
		list.SetMainTextColor(tview.Styles.PrimaryTextColor)
		list.SetSelectedTextColor(tview.Styles.InverseTextColor)
		list.SetSelectedBackgroundColor(tview.Styles.SecondaryTextColor)
		list.SetSecondaryTextColor(tview.Styles.SecondaryTextColor)
		list.SetShortcutColor(tview.Styles.TertiaryTextColor)
	}

	// Try to set text color (for TextView)
	if textView, ok := p.(*tview.TextView); ok {
		textView.SetTextColor(tview.Styles.PrimaryTextColor)
	}

	// Try to set field text color (for InputField)
	if inputField, ok := p.(*tview.InputField); ok {
		inputField.SetFieldTextColor(tview.Styles.PrimaryTextColor)
		inputField.SetFieldBackgroundColor(tview.Styles.ContrastBackgroundColor)
	}

	// Try to set label color (for various widgets)
	if labeled, ok := p.(interface{ SetLabelColor(tcell.Color) }); ok {
		labeled.SetLabelColor(tview.Styles.SecondaryTextColor)
	}
}
