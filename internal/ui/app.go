package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/config"
	"github.com/ramonvermeulen/whosthere/internal/discovery"
	"github.com/ramonvermeulen/whosthere/internal/discovery/arp"
	"github.com/ramonvermeulen/whosthere/internal/discovery/mdns"
	"github.com/ramonvermeulen/whosthere/internal/discovery/ssdp"
	"github.com/ramonvermeulen/whosthere/internal/oui"
	"github.com/ramonvermeulen/whosthere/internal/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/components"
	"github.com/ramonvermeulen/whosthere/internal/ui/navigation"
	"github.com/ramonvermeulen/whosthere/internal/ui/pages"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

const (
	// refreshInterval frequency of UI refreshes for redrawing tables/spinners/etc.
	refreshInterval = 1 * time.Second
)

// App is the public interface for running the TUI.
type App interface {
	Run() error
}

// tui is the concrete implementation of the App interface.
type tui struct {
	app          *tview.Application
	cfg          *config.Config
	router       *navigation.Router
	engine       *discovery.Engine
	state        *state.AppState
	version      string
	themeManager *theme.Manager
}

// NewApp constructs a new TUI instance using a builder-style initialization.
func NewApp(cfg *config.Config, ouiDB *oui.Registry, version string) App {
	app := tview.NewApplication()
	t := &tui{
		app:          app,
		cfg:          cfg,
		version:      version,
		themeManager: theme.NewManager(app),
	}

	return t.
		initializeTheme().
		buildEngine(ouiDB).
		buildState().
		buildRouter().
		buildPages().
		buildLayout().
		bindEvents()
}

// UIQueue returns a helper suitable for components that need to queue UI updates.
func (t *tui) UIQueue() func(func()) {
	return func(f func()) { t.app.QueueUpdateDraw(f) }
}

// initializeTheme resolves and applies the configured theme.
func (t *tui) initializeTheme() *tui {
	var themeName string
	if t.cfg != nil && t.cfg.Theme.Name != "" {
		themeName = t.cfg.Theme.Name
	} else {
		themeName = config.DefaultThemeName
	}

	// Set the initial theme through the manager
	t.themeManager.SetTheme(themeName)
	return t
}

// buildEngine constructs the discovery engine and scanners.
func (t *tui) buildEngine(ouiDB *oui.Registry) *tui {
	if t.cfg == nil {
		return t
	}

	sweeper := arp.NewSweeper(5*time.Minute, time.Minute)
	var scanners []discovery.Scanner

	if t.cfg.Scanners.SSDP.Enabled {
		scanners = append(scanners, &ssdp.Scanner{})
	}
	if t.cfg.Scanners.ARP.Enabled {
		scanners = append(scanners, arp.NewScanner(sweeper))
	}
	if t.cfg.Scanners.MDNS.Enabled {
		scanners = append(scanners, &mdns.Scanner{})
	}

	engine := discovery.NewEngine(
		scanners,
		discovery.WithTimeout(t.cfg.ScanDuration),
		discovery.WithOUIRegistry(ouiDB),
		discovery.WithSubnetHook(sweeper.Trigger),
	)

	t.engine = engine
	return t
}

// buildState initializes the shared application state store.
func (t *tui) buildState() *tui {
	t.state = state.NewAppState()
	return t
}

// buildRouter creates the navigation router.
func (t *tui) buildRouter() *tui {
	t.router = navigation.NewRouter()
	return t
}

// buildPages constructs and registers all pages with the router.
func (t *tui) buildPages() *tui {
	if t.router == nil {
		return t
	}

	dashboardPage := pages.NewDashboardPage(t.state, t.router.NavigateTo, t.version)
	detailPage := pages.NewDetailPage(t.state, t.router.NavigateTo, t.UIQueue(), t.version)
	splashPage := pages.NewSplashPage(t.version)

	t.router.Register(dashboardPage)
	t.router.Register(detailPage)
	t.router.Register(splashPage)

	return t
}

// buildLayout wires the router into the application root and sets the initial route.
func (t *tui) buildLayout() *tui {
	if t.router == nil {
		return t
	}

	if t.cfg != nil && t.cfg.Splash.Enabled {
		t.router.NavigateTo(navigation.RouteSplash)
	} else {
		t.router.NavigateTo(navigation.RouteDashboard)
	}

	t.app.SetRoot(t.router, true)
	t.router.FocusCurrent(t.app)
	return t
}

// bindEvents is a hook for global keybindings or input capture.
func (t *tui) bindEvents() *tui {
	// Set up global input capture for theme picker
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 't' {
			t.showThemePicker()
			return nil
		}
		return event
	})
	return t
}

// showThemePicker displays the theme picker modal.
func (t *tui) showThemePicker() {
	picker := components.NewThemePicker(t.themeManager)

	// Register the picker itself with the theme manager so it updates during preview
	t.themeManager.Register(picker)
	closeModal := func() {
		t.router.RemovePage("theme-picker-modal")
		t.router.FocusCurrent(t.app)
	}

	picker.OnSelect(func(themeName string) {
		// Theme is already applied during preview, just close the modal
		closeModal()
	})

	picker.OnSave(func(themeName string) {
		// Save the theme to config and close modal
		if err := t.saveThemeToConfig(themeName); err == nil {
			closeModal()
		}
	})

	picker.OnCancel(func() {
		// Close modal (theme restoration handled by picker)
		closeModal()
	})

	picker.Show()

	// Create a centered modal layout with semi-transparent background effect
	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(picker.GetList(), 80, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	// Add as an overlay page on top of current view
	t.router.AddPage("theme-picker-modal", modal, true, true)
	t.app.SetFocus(picker.GetListPrimitive())
}

// saveThemeToConfig updates the config with the new theme and saves it to disk.
func (t *tui) saveThemeToConfig(themeName string) error {
	if t.cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	// Update the theme name in the config
	t.cfg.Theme.Name = themeName

	// Save the config to disk
	if err := config.Save(t.cfg, ""); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Run starts the TUI event loop and background workers.
func (t *tui) Run() error {

	if t.cfg != nil && t.cfg.Splash.Enabled {
		go func(delay time.Duration) {
			time.Sleep(delay)
			t.app.QueueUpdateDraw(func() {
				if t.router != nil {
					t.router.NavigateTo(navigation.RouteDashboard)
					t.router.FocusCurrent(t.app)
				}
				t.startBackgroundTasks()
			})
		}(t.cfg.Splash.Delay)
	} else {
		t.startBackgroundTasks()
	}
	return t.app.Run()
}

// startBackgroundTasks launches app-wide background workers (UI refresh, discovery scanning).
func (t *tui) startBackgroundTasks() {
	t.startDashboardRefreshLoop()
	t.startDiscoveryScanLoop()
}

// startDashboardRefreshLoop periodically refreshes the dashboard view from state.
func (t *tui) startDashboardRefreshLoop() {
	if t.router == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			if t.router.Current() != navigation.RouteDashboard {
				continue
			}
			mp, _ := t.router.Page(navigation.RouteDashboard).(*pages.DashboardPage)
			if mp == nil {
				continue
			}
			t.app.QueueUpdateDraw(func() { mp.RefreshFromState() })
		}
	}()
}

// startDiscoveryScanLoop runs periodic network discovery and controls the spinner around scans.
func (t *tui) startDiscoveryScanLoop() {
	if t.cfg == nil || t.engine == nil || t.router == nil || t.state == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(t.cfg.ScanInterval)
		defer ticker.Stop()

		doScan := func() {
			mp, _ := t.router.Page(navigation.RouteDashboard).(*pages.DashboardPage)
			if mp == nil {
				return
			}
			mp.Spinner().Start(t.UIQueue())
			ctx := context.Background()
			cctx, cancel := context.WithTimeout(ctx, t.cfg.ScanDuration)
			_, _ = t.engine.Stream(cctx, func(d discovery.Device) {
				t.state.UpsertDevice(&d)
			})
			cancel()
			mp.Spinner().Stop(t.UIQueue())
		}

		doScan()

		for range ticker.C {
			doScan()
		}
	}()
}
