package pages

import (
	"fmt"
	"strings"

	"github.com/derailed/tview"
)

var LogoBig = []string{
	`Knock Knock..                                                     `,
	`                _               _   _                   ___       `,
	`      __      _| |__   ___  ___| |_| |__   ___ _ __ ___/ _ \      `,
	`      \ \ /\ / / '_ \ / _ \/ __| __| '_ \ / _ \ '__/ _ \// /      `,
	`       \ V  V /| | | | (_) \__ \ |_| | | |  __/ | |  __/ \/       `,
	`        \_/\_/ |_| |_|\___/|___/\__|_| |_|\___|_|  \___| ()       `,
}

// SplashPage adapts the splash logo into a Page.
type SplashPage struct {
	root *tview.Flex
}

func NewSplashPage() *SplashPage {
	s := &SplashPage{root: tview.NewFlex()}

	logo := tview.NewTextView()
	logo.SetDynamicColors(true)
	logo.SetTextAlign(tview.AlignCenter)

	// TODO(ramon): fix styles via injectable style configuration
	logoText := strings.Join(LogoBig, "\n[green::b]")
	_, _ = fmt.Fprintf(logo, "%s[green::b]%s\n",
		strings.Repeat("\n", 2),
		logoText)

	s.root.AddItem(logo, 0, 1, false)

	return s
}

func (p *SplashPage) GetName() string { return "splash" }

func (p *SplashPage) GetPrimitive() tview.Primitive { return p.root }

func (p *SplashPage) FocusTarget() tview.Primitive { return p.root }
