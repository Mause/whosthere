package ui

import (
	"time"

	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/config"
)

type App struct {
	*tview.Application
	Main *Pages
	cfg  *config.Config
}

func NewApp(cfg *config.Config) *App {
	a := App{
		Application: tview.NewApplication(),
		Main:        NewPages(),
		cfg:         cfg,
	}

	return &a
}

func (a *App) Init() error {
	a.layout()
	return nil
}

func (a *App) Run() error {
	if a.cfg != nil && a.cfg.Splash.Enabled {
		go func(delaySeconds int) {
			timer := time.NewTimer(time.Duration(delaySeconds) * time.Second)
			<-timer.C

			a.QueueUpdateDraw(func() {
				a.Main.SwitchToPage("main")
			})
		}(a.cfg.Splash.Delay)
	} else {
		a.Main.SwitchToPage("main")
	}

	if err := a.Application.Run(); err != nil {
		return err
	}

	return nil
}

func (a *App) layout() {
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(tview.NewTextView().SetText("whosthere").SetTextAlign(tview.AlignCenter), 0, 1, false)
	main.AddItem(tview.NewBox().SetTitle("Devices").SetBorder(true), 0, 18, true)
	main.AddItem(tview.NewTextView().SetText("jK up/down - gG top/bottom").SetTextAlign(tview.AlignCenter), 0, 1, false)

	a.Main.AddPage("main", main, true, false)
	a.Main.AddPage("splash", NewSplash(), true, true)

	a.SetRoot(a.Main, true)
}
