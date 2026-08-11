package locationinfo

import (
	"farental/core/data/api"
	ftheme "farental/internal/theme"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
)

func TestMain(m *testing.M) {
	orvyn.Init()
	orvyn.SetTheme(ftheme.NewFarentalDarkTheme())

	os.Exit(m.Run())
}

const descriptionText = "A quiet harbour town built on the ruins of an older " +
	"port, where the tide still uncovers foundations no one remembers laying."

func testLocation() api.LocationResponse {
	return api.LocationResponse{
		Name:            "Port Halden",
		Description:     "A harbour town.",
		LongDescription: descriptionText,
		Continent:       api.LocationInfoResponse{Name: "Aldoria", Description: "The western continent."},
		Type:            api.LocationInfoResponse{Name: "Town", Description: "A settled place."},
		Biome:           api.LocationInfoResponse{Name: "Beach", Description: "Sand and salt."},
		Features: []api.LocationFeatureResponse{
			{LocationInfoResponse: api.LocationInfoResponse{
				Code: "shop", Name: "Shop", Description: "Buy and sell goods."}},
			{LocationInfoResponse: api.LocationInfoResponse{
				Code: "bank", Name: "Bank", Description: "Store your coin."}},
		},
	}
}

// renderedScreen lays a screen out at the given terminal size and returns it,
// so the individual widgets can be measured after the layout has sized them.
func renderedScreen(t *testing.T, width, height int) *Screen {
	t.Helper()

	s := New()

	location := testLocation()
	s.updateData(&location)

	s.layout.Resize(orvyn.NewSize(width, height))
	s.layout.Render()

	return s
}

// renderAt lays the screen out at the given terminal size and returns the view.
func renderAt(t *testing.T, width, height int) string {
	t.Helper()

	s := New()

	location := testLocation()
	s.updateData(&location)

	s.layout.Resize(orvyn.NewSize(width, height))

	return s.layout.Render()
}

// The description viewer was allocated a single row, which its border consumed
// whole, so the screen showed a titled but empty box.
func TestDescriptionIsRendered(t *testing.T) {
	view := renderAt(t, 120, 40)

	// The description wraps, so only the opening words survive as one string.
	if !strings.Contains(view, "A quiet harbour town") {
		t.Errorf("the description is missing from the rendered screen:\n%s", view)
	}
}

// With every child reporting a fixed height, the layout left most of the
// terminal empty instead of sharing it between the description and the features
// row. The centering layout pads the view back to the terminal height whatever
// happens, so the block heights are what has to be measured.
func TestLeftoverHeightIsShared(t *testing.T) {
	const short, tall = 30, 60

	shortScreen := renderedScreen(t, 120, short)
	tallScreen := renderedScreen(t, 120, tall)

	shortDescription := lipgloss.Height(shortScreen.description.Render())
	tallDescription := lipgloss.Height(tallScreen.description.Render())

	if tallDescription <= shortDescription {
		t.Errorf("description height = %d rows in a %d-row terminal and %d in a "+
			"%d-row one; it must grow with the terminal",
			shortDescription, short, tallDescription, tall)
	}

	shortFeatures := lipgloss.Height(shortScreen.featuresList.Render())
	tallFeatures := lipgloss.Height(tallScreen.featuresList.Render())

	if tallFeatures <= shortFeatures {
		t.Errorf("features height = %d rows in a %d-row terminal and %d in a "+
			"%d-row one; it must grow with the terminal",
			shortFeatures, short, tallFeatures, tall)
	}
}

// The whole screen must stay inside the terminal: orvyn clips the overflow off
// the bottom, which silently swallows the help bar.
func TestScreenDoesNotOverflow(t *testing.T) {
	for _, height := range []int{24, 40, 60} {
		view := renderAt(t, 120, height)

		if got := lipgloss.Height(view); got > height {
			t.Errorf("rendered height = %d overflows a %d-row terminal:\n%s",
				got, height, view)
		}
	}
}

// The features and the cards must both survive, at the smallest height the
// screen is expected to cope with as well as at a roomy one.
func TestFeaturesAndCardsSurvive(t *testing.T) {
	for _, height := range []int{24, 40, 60} {
		view := renderAt(t, 120, height)

		for _, want := range []string{"Shop", "Bank", "Aldoria", "Town", "Beach"} {
			if !strings.Contains(view, want) {
				t.Errorf("height %d: %q is missing from the rendered screen:\n%s",
					height, want, view)
			}
		}
	}
}

// A location with no long description falls back to the short one, which must
// reach the viewer just the same.
func TestShortDescriptionFallback(t *testing.T) {
	s := New()

	location := testLocation()
	location.LongDescription = ""

	s.updateData(&location)

	s.layout.Resize(orvyn.NewSize(120, 40))

	if view := s.layout.Render(); !strings.Contains(view, "A harbour town.") {
		t.Errorf("the short description is missing from the rendered screen:\n%s",
			view)
	}
}
