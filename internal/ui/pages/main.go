package pages

import (
	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
)

// MainPage is the dashboard showing discovered devices.
type MainPage struct {
	root        *tview.Flex
	deviceTable *components.DeviceTable
	spinner     *components.Spinner
	state       *state.AppState

	onShowDetails func()
}

func NewMainPage(s *state.AppState, onShowDetails func()) *MainPage {
	t := components.NewDeviceTable()
	spinner := components.NewSpinner()

	mp := &MainPage{
		root:          nil, // set below
		deviceTable:   t,
		spinner:       spinner,
		state:         s,
		onShowDetails: onShowDetails,
	}

	t.SetSelectedFunc(func(row, col int) {
		ip := mp.deviceTable.SelectedIP()
		if ip == "" || mp.onShowDetails == nil {
			return
		}
		mp.state.SetSelectedIP(ip)
		mp.onShowDetails()
	})

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(tview.NewTextView().SetText("whosthere").SetTextAlign(tview.AlignCenter), 0, 1, false)
	main.AddItem(t, 0, 18, true)

	status := tview.NewFlex().SetDirection(tview.FlexColumn)
	status.AddItem(mp.spinner.View(), 0, 1, false)
	status.AddItem(tview.NewTextView().SetText("j/k: up/down  g/G: top/bottom  Enter: details").SetTextAlign(tview.AlignRight), 0, 1, false)
	main.AddItem(status, 1, 0, false)

	mp.root = main
	return mp
}

func (p *MainPage) GetName() string { return "main" }

func (p *MainPage) GetPrimitive() tview.Primitive { return p.root }

func (p *MainPage) FocusTarget() tview.Primitive { return p.deviceTable }

func (p *MainPage) Spinner() *components.Spinner { return p.spinner }

func (p *MainPage) RefreshFromState() {
	devices := p.state.DevicesSnapshot()
	p.deviceTable.ReplaceAll(devices)
}
