package pages

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/state"
)

// DetailPage shows detailed information about the currently selected device.
type DetailPage struct {
	root  *tview.Flex
	state *state.AppState
	info  *tview.TextView

	// onBack is called when the user wants to return to the main page (q or Esc).
	onBack func()
}

func NewDetailPage(s *state.AppState, onBack func()) *DetailPage {
	info := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	info.SetBorder(true).SetTitle("Device details (q/Esc to go back)")

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(info, 0, 1, true)

	p := &DetailPage{
		root:   root,
		state:  s,
		info:   info,
		onBack: onBack,
	}

	info.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev == nil {
			return ev
		}
		if ev.Rune() == 'q' || ev.Key() == tcell.KeyEsc {
			if p.onBack != nil {
				p.onBack()
			}
			return nil
		}
		return ev
	})

	p.Refresh()
	return p
}

func (p *DetailPage) GetName() string { return "detail" }

func (p *DetailPage) GetPrimitive() tview.Primitive { return p.root }

// FocusTarget returns the main text view so it receives input (for q/Esc) when this page is active.
func (p *DetailPage) FocusTarget() tview.Primitive { return p.info }

// Refresh reloads the text view from the currently selected device, if any.
func (p *DetailPage) Refresh() {
	p.info.Clear()
	d, ok := p.state.Selected()
	if !ok {
		_, _ = fmt.Fprintln(p.info, "No device selected.")
		return
	}

	_, _ = fmt.Fprintf(p.info, "[yellow::b]IP:[-::-] %s\n", d.IP)
	_, _ = fmt.Fprintf(p.info, "[yellow::b]Hostname:[-::-] %s\n", d.Hostname)
	_, _ = fmt.Fprintf(p.info, "[yellow::b]MAC:[-::-] %s\n", d.MAC)
	_, _ = fmt.Fprintf(p.info, "[yellow::b]Manufacturer:[-::-] %s\n", d.Manufacturer)
	_, _ = fmt.Fprintf(p.info, "[yellow::b]Model:[-::-] %s\n\n", d.Model)

	_, _ = fmt.Fprintf(p.info, "[yellow::b]Services:[-::-]\n")
	if len(d.Services) == 0 {
		_, _ = fmt.Fprintln(p.info, "  (none)")
	} else {
		for name, port := range d.Services {
			_, _ = fmt.Fprintf(p.info, "  %s:%d\n", name, port)
		}
	}

	_, _ = fmt.Fprintf(p.info, "\n[yellow::b]Sources:[-::-]\n")
	if len(d.Sources) == 0 {
		_, _ = fmt.Fprintln(p.info, "  (none)")
	} else {
		for src := range d.Sources {
			_, _ = fmt.Fprintf(p.info, "  %s\n", src)
		}
	}
}
