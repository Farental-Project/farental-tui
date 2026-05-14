package travel

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/generic/selectionlist"
	"farental/widget/travellistitem"
	"farental/widget/travelrelaylistitem"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

type mode int

const (
	modeTravelSelection mode = iota
	modeTravelRelaySelection
)

type Screen struct {
	travelSelectionScreen      selectionlist.Screen[api.TravelResponse]
	travelRelaySelectionScreen selectionlist.Screen[api.TravelRelayResponse]
	currentMode                mode
}

func New() *Screen {
	s := new(Screen)

	s.travelSelectionScreen = selectionlist.New(lokyn.L("Travels"), travellistitem.Constructor,
		s.loadTravels, s.submit)
	s.travelRelaySelectionScreen = selectionlist.New(lokyn.L("Relays"), travelrelaylistitem.Constructor,
		s.loadTravelRelays, s.submitRelay)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.currentMode = modeTravelSelection

	s.travelSelectionScreen.OnEnter(i)
	s.travelSelectionScreen.SetTitle(lokyn.L("Travels"))
	s.travelRelaySelectionScreen.OnEnter(i)
	s.travelRelaySelectionScreen.SetTitle(lokyn.L("Relays"))

	bubblehelp.SwitchContext(keybind.ContextTravel)

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Tab):
			switch s.currentMode {
			case modeTravelSelection:
				s.currentMode = modeTravelRelaySelection
				bubblehelp.UpdateKeybindHelpDesc(keybind.Tab, lokyn.L("travels"))
				return nil
			case modeTravelRelaySelection:
				s.currentMode = modeTravelSelection
				bubblehelp.UpdateKeybindHelpDesc(keybind.Tab, lokyn.L("")) // Default
				return nil
			}
		}
	}

	cmd = s.getCurrentScreen().Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.getCurrentScreen().Render()
}

func (s *Screen) submit() bool {
	i := s.travelSelectionScreen.GetSelectedItem()

	req := request.TravelStart(i.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.travelSelectionScreen.SetStatusError(err)
		return false
	}

	if resp.StatusCode() != 200 {
		return false
	}

	return true
}

func (s *Screen) loadTravels() {
	var travels *[]api.TravelResponse
	var ok bool

	resp, err := helper.SendRequest(request.TravelGetAvailable())

	if err != nil {
		s.travelSelectionScreen.SetStatusError(err)
		return
	}

	travels, ok = resp.Result().(*[]api.TravelResponse)

	if !ok {
		return
	}

	s.travelSelectionScreen.SetItems(*travels)
}

func (s *Screen) submitRelay() bool {
	i := s.travelRelaySelectionScreen.GetSelectedItem()

	req := request.TravelRelayStart(i.DestLocation.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.travelRelaySelectionScreen.SetStatusError(err)
		return false
	}

	if resp.StatusCode() != 200 {
		return false
	}

	return true
}

func (s *Screen) loadTravelRelays() {
	var travels *[]api.TravelRelayResponse
	var ok bool

	resp, err := helper.SendRequest(request.TravelRelayGetAvailable())

	if err != nil {
		s.travelRelaySelectionScreen.SetStatusError(err)
		return
	}

	travels, ok = resp.Result().(*[]api.TravelRelayResponse)

	if !ok {
		return
	}

	s.travelRelaySelectionScreen.SetItems(*travels)
}

func (s *Screen) getCurrentScreen() orvyn.Screen {
	switch s.currentMode {
	case modeTravelSelection:
		return &s.travelSelectionScreen
	case modeTravelRelaySelection:
		return &s.travelRelaySelectionScreen
	}

	return nil
}
