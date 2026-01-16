package components

import (
	"time"

	"github.com/ramonvermeulen/whosthere/internal/core/state"
	"github.com/ramonvermeulen/whosthere/internal/ui/theme"
	"github.com/rivo/tview"
)

var _ UIComponent = &Spinner{}

type Spinner struct {
	*tview.TextView
	stop    chan struct{}
	running bool
	suffix  string
}

func NewSpinner() *Spinner {
	tv := tview.NewTextView().SetText(" ").SetTextAlign(tview.AlignLeft)
	theme.RegisterPrimitive(tv)
	return &Spinner{TextView: tv, stop: make(chan struct{}, 1), suffix: ""}
}

func (s *Spinner) SetSuffix(suf string) { s.suffix = suf }

// Start runs the spinner loop and uses the provided queue function to schedule UI updates.
func (s *Spinner) Start(queue func(f func())) {
	if s.running {
		return
	}
	s.running = true

	frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	interval := 100 * time.Millisecond

	select {
	case <-s.stop:
	default:
	}

	go func() {
		idx := 0
		for {
			select {
			case <-s.stop:
				s.running = false
				queue(func() { s.SetText("") })
				return
			case <-time.After(interval):
				ch := string(frames[idx%len(frames)])
				idx++
				queue(func() { s.SetText(ch + s.suffix) })
			}
		}
	}()
}

// Stop signals the spinner goroutine to stop and schedules a final clear.
func (s *Spinner) Stop(queue func(f func())) {
	select {
	case s.stop <- struct{}{}:
	default:
	}
	queue(func() { s.SetText("") })
	s.running = false
}

// Render implements UIComponent.
func (s *Spinner) Render(_ state.ReadOnly) {
	// Spinner is animated separately, no state update needed.
}
