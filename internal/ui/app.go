package ui

import (
	"context"
	"math"
	"time"

	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/config"
	"github.com/ramonvermeulen/whosthere/internal/discovery"
	"github.com/ramonvermeulen/whosthere/internal/discovery/arp"
	"github.com/ramonvermeulen/whosthere/internal/discovery/mdns"
	"github.com/ramonvermeulen/whosthere/internal/discovery/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/oui"
	"github.com/ramonvermeulen/whosthere/internal/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
	"github.com/ramonvermeulen/whosthere/internal/ui/pages"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
)

const (
	// refreshInterval frequency of UI refreshes for redrawing tables/spinners/etc.
	refreshInterval = 1 * time.Second
)

type App struct {
	*tview.Application

	cfg    *config.Config
	router *navigation.Router
	engine *discovery.Engine
	state  *state.AppState
}

func (a *App) UIQueue() func(func()) {
	return func(f func()) { a.QueueUpdateDraw(f) }
}

func NewApp(cfg *config.Config, ouiDB *oui.Registry) *App {
	_ = theme.FromConfig(cfg.Theme)
	scanners := []discovery.Scanner{
		&ssdp.Scanner{},
		&arp.Scanner{},
		&mdns.Scanner{},
	}
	engine := discovery.NewEngine(
		scanners,
		discovery.WithTimeout(cfg.ScanDuration),
		discovery.WithOUIRegistry(ouiDB),
	)

	a := &App{
		Application: tview.NewApplication(),
		cfg:         cfg,
		router:      navigation.NewRouter(),
		engine:      engine,
		state:       state.NewAppState(),
	}

	dashboardPage := pages.NewDashboardPage(a.state, a.router.NavigateTo)
	detailPage := pages.NewDetailPage(a.state, a.router.NavigateTo, a.UIQueue())
	splashPage := pages.NewSplashPage()

	a.router.Register(dashboardPage)
	a.router.Register(detailPage)
	a.router.Register(splashPage)

	if a.cfg != nil && a.cfg.Splash.Enabled {
		a.router.NavigateTo(navigation.RouteSplash)
	} else {
		a.router.NavigateTo(navigation.RouteDashboard)
	}

	a.SetRoot(a.router, true)
	a.router.FocusCurrent(a.Application)

	return a
}

func (a *App) Run() error {
	if a.cfg != nil && a.cfg.Splash.Enabled {
		go func(delaySeconds float32) {
			ms := int64(math.Round(float64(delaySeconds) * 1000.0))
			time.Sleep(time.Duration(ms) * time.Millisecond)
			a.QueueUpdateDraw(func() {
				a.router.NavigateTo(navigation.RouteDashboard)
				a.router.FocusCurrent(a.Application)
				a.startBackgroundTasks()
			})
		}(a.cfg.Splash.Delay)
	} else {
		a.startBackgroundTasks()
	}
	return a.Application.Run()
}

// startBackgroundTasks launches app-wide background workers (UI refresh, discovery scanning).
func (a *App) startBackgroundTasks() {
	a.startDashboardRefreshLoop()
	a.startDiscoveryScanLoop()
}

// startDashboardRefreshLoop periodically refreshes the dashboard view from state.
func (a *App) startDashboardRefreshLoop() {
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			if a.router.Current() != navigation.RouteDashboard {
				continue
			}
			mp, _ := a.router.Page(navigation.RouteDashboard).(*pages.DashboardPage)
			if mp == nil {
				continue
			}
			a.QueueUpdateDraw(func() { mp.RefreshFromState() })
		}
	}()
}

// startDiscoveryScanLoop runs periodic network discovery and controls the spinner around scans.
func (a *App) startDiscoveryScanLoop() {
	go func() {
		ticker := time.NewTicker(a.cfg.ScanInterval)
		defer ticker.Stop()

		doScan := func() {
			mp, _ := a.router.Page(navigation.RouteDashboard).(*pages.DashboardPage)
			if mp == nil {
				return
			}
			mp.Spinner().Start(a.UIQueue())
			ctx := context.Background()
			cctx, cancel := context.WithTimeout(ctx, a.cfg.ScanDuration)
			_, _ = a.engine.Stream(cctx, func(d discovery.Device) {
				a.state.UpsertDevice(d)
			})
			cancel()
			mp.Spinner().Stop(a.UIQueue())
		}

		doScan()

		for range ticker.C {
			doScan()
		}
	}()
}
