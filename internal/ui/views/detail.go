package views

import (
	"fmt"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/events"
	"github.com/ramonvermeulen/whosthere/internal/ui/routes"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ View = &DetailView{}

// DetailView shows detailed information about the currently selected device.
type DetailView struct {
	*tview.Flex
	info      *tview.TextView
	header    *components.Header
	statusBar *components.StatusBar

	emit func(events.Event)
}

func NewDetailView(emit func(events.Event)) *DetailView {
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	header := components.NewHeader()

	info := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	info.SetBorder(true).
		SetTitle("Details").
		SetTitleColor(tview.Styles.TitleColor).
		SetBorderColor(tview.Styles.BorderColor).
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	statusBar := components.NewStatusBar()
	statusBar.SetHelp("Esc/q: Back")

	main.AddItem(header, 1, 0, false)
	main.AddItem(info, 0, 1, true)
	main.AddItem(statusBar, 1, 0, false)

	p := &DetailView{
		Flex:      main,
		info:      info,
		header:    header,
		statusBar: statusBar,
		emit:      emit,
	}

	info.SetInputCapture(handleInput(p))

	theme.RegisterPrimitive(p)
	theme.RegisterPrimitive(p.info)

	return p
}

func handleInput(p *DetailView) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		if ev == nil {
			return nil
		}
		switch {
		case ev.Key() == tcell.KeyEsc || ev.Rune() == 'q':
			p.emit(events.NavigateTo{Route: routes.RouteDashboard, Overlay: true})
			return nil
		default:
			return ev
		}
	}
}

func (d *DetailView) FocusTarget() tview.Primitive { return d.info }

// Render reloads the text view from the currently selected device, if any.
func (d *DetailView) Render(s state.ReadOnly) {
	d.info.Clear()
	device, ok := s.Selected()
	if !ok {
		_, _ = fmt.Fprintln(d.info, "No device selected.")
		return
	}

	labelColor := colorToHexTag(tview.Styles.SecondaryTextColor)
	valueColor := colorToHexTag(tview.Styles.PrimaryTextColor)
	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	}

	_, _ = fmt.Fprintf(d.info, "[%s::b]IP:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, device.IP)
	_, _ = fmt.Fprintf(d.info, "[%s::b]Display Name:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, device.DisplayName)
	_, _ = fmt.Fprintf(d.info, "[%s::b]MAC:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, device.MAC)
	_, _ = fmt.Fprintf(d.info, "[%s::b]Manufacturer:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, device.Manufacturer)
	_, _ = fmt.Fprintf(d.info, "[%s::b]First Seen:[-::-] [%s::]%s[-::-]\n", labelColor, valueColor, formatTime(device.FirstSeen))
	_, _ = fmt.Fprintf(d.info, "[%s::b]Last Seen:[-::-] [%s::]%s[-::-]\n\n", labelColor, valueColor, formatTime(device.LastSeen))

	_, _ = fmt.Fprintf(d.info, "[%s::b]Sources:[-::-]\n", labelColor)
	if len(device.Sources) == 0 {
		_, _ = fmt.Fprintln(d.info, "  (none)")
	} else {
		for _, src := range sortedKeys(device.Sources) {
			_, _ = fmt.Fprintf(d.info, "  %s\n", src)
		}
	}

	_, _ = fmt.Fprintf(d.info, "\n[%s::b]Open Ports:[-::-]\n", labelColor)
	if len(device.OpenPorts) == 0 {
		_, _ = fmt.Fprintln(d.info, "  (none)")
	} else {
		for _, port := range device.OpenPorts {
			_, _ = fmt.Fprintf(d.info, "  %device\n", port)
		}
		if !device.LastPortScan.IsZero() {
			_, _ = fmt.Fprintf(d.info, "\n[%s::b]Last portscan:[-::-] %s\n", labelColor, device.LastPortScan.Format("2006-01-02 15:04:05"))
		}
	}

	_, _ = fmt.Fprintf(d.info, "\n[%s::b]Extra Data:[-::-]\n", labelColor)
	if len(device.ExtraData) == 0 {
		_, _ = fmt.Fprintln(d.info, "  (none)")
	} else {
		for _, k := range sortedKeys(device.ExtraData) {
			_, _ = fmt.Fprintf(d.info, "  %s: %s\n", k, device.ExtraData[k])
		}
	}

	d.header.Render(s)
	d.statusBar.Render(s)
}

// colorToHexTag converts a tcell.Color to a tview dynamic color hex tag.
func colorToHexTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// sortedKeys is a helper to return asc sorted map keys.
func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
