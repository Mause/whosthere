package ui

import (
	"github.com/derailed/tview"
	"github.com/ramonvermeulen/whosthere/internal/ui/pages"
)

// Route names for top-level application pages.
const (
	RouteMain   = "main"
	RouteSplash = "splash"
	RouteDetail = "detail"
)

// Router is both the visual pages container and the logical router
// that knows about Page implementations and current page state.
type Router struct {
	*tview.Pages

	pages       map[string]pages.Page
	currentPage string
}

func NewRouter() *Router {
	return &Router{
		Pages: tview.NewPages(),
		pages: make(map[string]pages.Page),
	}
}

// Register adds a Page and attaches its primitive to the underlying tview.Pages.
func (r *Router) Register(p pages.Page) {
	name := p.GetName()
	r.pages[name] = p
	r.AddPage(name, p.GetPrimitive(), true, false)
}

// NavigateTo switches to a previously registered page by name.
func (r *Router) NavigateTo(name string) {
	if _, ok := r.pages[name]; !ok {
		return
	}
	r.currentPage = name
	r.SwitchToPage(name)
}

// FocusCurrent sets focus to the current page's preferred primitive.
func (r *Router) FocusCurrent(app *tview.Application) {
	if app == nil {
		return
	}
	p, ok := r.pages[r.currentPage]
	if !ok || p == nil {
		app.SetFocus(r)
		return
	}
	if ft := p.FocusTarget(); ft != nil {
		app.SetFocus(ft)
		return
	}
	app.SetFocus(p.GetPrimitive())
}

// Current returns the currently active page name.
func (r *Router) Current() string { return r.currentPage }

// Page returns a registered page by name.
func (r *Router) Page(name string) pages.Page {
	return r.pages[name]
}
