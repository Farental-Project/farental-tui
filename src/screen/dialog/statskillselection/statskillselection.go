package statskillselection

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	"farental/screen/generic/selectionlist"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type EntryKind int

const (
	// EntryNone clears the stat filter.
	EntryNone EntryKind = iota
	EntryStat
	EntrySkill
)

// Entry is one pickable value. A skill stands for its primordial stat, which
// is why stats and skills share a single control.
type Entry struct {
	Kind  EntryKind
	Code  string
	Label string
}

type Screen struct {
	selectionlist.Screen[Entry]

	options *api.AuctionFilterOptionsResponse

	submitted bool
}

var _ orvyn.Screen = (*Screen)(nil)

func New(options *api.AuctionFilterOptionsResponse) *Screen {
	s := new(Screen)

	s.options = options

	s.Screen = selectionlist.New(lokyn.L("Stat or skill"),
		Constructor, s.loadData, s.submit)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.submitted = false

	cmd := s.Screen.OnEnter(i)

	s.SetTitle(lokyn.L("Stat or skill"))

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListBasic)

	return cmd
}

func (s *Screen) OnExit() any {
	if s.submitted {
		return s.GetSelectedItem()
	}

	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Enter):
			if s.GetFilteringState() != widgetlist.Filtering {
				s.submitted = true
				return orvyn.CloseDialog()
			}

		case key.Matches(m, keybind.Esc):
			if s.GetFilteringState() == widgetlist.Unfiltered {
				return orvyn.CloseDialog()
			}
		}
	}

	return s.Screen.Update(msg)
}

func (s *Screen) loadData() {
	entries := []Entry{{Kind: EntryNone, Label: lokyn.L("Any stat or skill")}}

	if s.options != nil {
		for _, stat := range s.options.Stats {
			entries = append(entries, Entry{
				Kind:  EntryStat,
				Code:  stat.Code,
				Label: stat.Name,
			})
		}

		for _, skill := range s.options.Skills {
			entries = append(entries, Entry{
				Kind:  EntrySkill,
				Code:  skill.Code,
				Label: skill.Name,
			})
		}
	}

	s.SetItems(entries)
}

// submit satisfies selectionlist's callback; the dialog closes from Update,
// which is where it knows it is a dialog and not a screen.
func (s *Screen) submit() bool {
	return true
}
