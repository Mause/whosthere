package ui

import (
	"context"
	"math"
	"time"

	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/config"
	"github.com/ramonvermeulen/whosthere/internal/discovery"
	"github.com/ramonvermeulen/whosthere/internal/discovery/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/pages"
)

type App struct {
	*tview.Application

	cfg    *config.Config
	router *Router
	engine *discovery.Engine
	state  *state.AppState
}

func NewApp(cfg *config.Config) *App {
	a := &App{
		Application: tview.NewApplication(),
		cfg:         cfg,
		router:      NewRouter(),
		engine:      &discovery.Engine{Scanners: []discovery.Scanner{&ssdp.Scanner{}}, Timeout: 6 * time.Second},
		state:       state.NewAppState(),
	}

	detailPage := pages.NewDetailPage(a.state, func() {
		a.router.NavigateTo(RouteMain)
	})
	mainPage := pages.NewMainPage(a.state, func() {
		if dp, ok := a.router.Page(RouteDetail).(*pages.DetailPage); ok {
			dp.Refresh()
		}
		a.router.NavigateTo(RouteDetail)
	})
	splashPage := pages.NewSplashPage()

	a.router.Register(mainPage)
	a.router.Register(detailPage)
	a.router.Register(splashPage)

	if a.cfg != nil && a.cfg.Splash.Enabled {
		a.router.NavigateTo(RouteSplash)
	} else {
		a.router.NavigateTo(RouteMain)
	}

	a.SetRoot(a.router, true)
	// Ensure focus is on the active page's focus target (e.g., the device table on main page).
	a.router.FocusCurrent(a.Application)

	return a
}

func (a *App) Run() error {
	if a.cfg != nil && a.cfg.Splash.Enabled {
		go func(delaySeconds float32) {
			ms := int64(math.Round(float64(delaySeconds) * 1000.0))
			timer := time.NewTimer(time.Duration(ms) * time.Millisecond)
			<-timer.C
			a.QueueUpdateDraw(func() {
				a.router.NavigateTo(RouteMain)
				// After navigating away from the splash, focus the main page's focus target (device table).
				a.router.FocusCurrent(a.Application)
			})
			a.startDiscoveryLoop()
		}(a.cfg.Splash.Delay)
	} else {
		a.startDiscoveryLoop()
	}
	return a.Application.Run()
}

func (a *App) startDiscoveryLoop() {
	queue := func(f func()) { a.QueueUpdateDraw(f) }

	go func() {
		for {
			mp, _ := a.router.Page(RouteMain).(*pages.MainPage)
			if mp == nil {
				return
			}

			mp.Spinner().Start(queue)

			ctx := context.Background()
			_, _ = a.engine.Stream(ctx, func(d discovery.Device) {
				a.state.UpsertDevice(d)
				a.QueueUpdateDraw(func() { mp.RefreshFromState() })
			})

			mp.Spinner().Stop(queue)
			a.QueueUpdateDraw(func() { mp.RefreshFromState() })

			time.Sleep(10 * time.Second)
		}
	}()
}
