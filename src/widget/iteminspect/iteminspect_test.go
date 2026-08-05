package iteminspect

import (
	"farental/core/data/api"
	ftheme "farental/internal/theme"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

func TestMain(m *testing.M) {
	lokyn.Init()
	lokyn.SetLanguage("en")

	orvyn.SetTheme(ftheme.NewFarentalDarkTheme())

	os.Exit(m.Run())
}

// testItem is deliberately busy: a long description plus equipment stats and
// conditions, so every optional section of the widget renders. The
// description has no literal newlines, so GetMinSize (which measures the
// unwrapped string) will under-report its height once it wraps at a narrow
// width - which is exactly the gap this fix has to cover.
func testItem() *api.ItemResponse {
	stats := []api.EquipmentStatResponse{
		{Stat: api.StatResponse{Name: "Strength"}, Value: 12},
		{Stat: api.StatResponse{Name: "Agility"}, Value: 8},
		{Stat: api.StatResponse{Name: "Vitality"}, Value: 5},
	}

	conditions := []string{
		"Level 10 or higher",
		"Strength 15 or higher",
		"Guild rank Adept or higher",
	}

	return &api.ItemResponse{
		ID:   1,
		Name: "Blade of Endless Testing",
		Description: strings.Repeat(
			"A long and winding description meant to wrap across many lines "+
				"when the widget is only given a narrow column to render into. ", 6),
		IsUnique:       true,
		MaxStackCount:  1,
		EquipmentSlot:  &api.EquipmentSlotResponse{Code: "weapon", Name: "Weapon"},
		EquipmentStats: &stats,
		Conditions:     &conditions,
	}
}

// TestRenderNeverExceedsAllocatedHeight is the contract the fix establishes:
// whatever height the widget is resized to, Render must produce exactly that
// many lines - never more (the overflow this fix closes, since lipgloss's
// Height only pads and VBoxLayout renders children at their natural size
// regardless of what height it was allocated) and never fewer (Height still
// pads a short render up to the allocation).
func TestRenderNeverExceedsAllocatedHeight(t *testing.T) {
	const width = 30

	// Measure the content's true rendered height at this width, unclipped:
	// a generous allocation reports the layout's natural size as-is, which is
	// larger than GetMinSize predicts because the description wraps at this
	// width.
	probe := New()
	probe.UpdateData(testItem())
	probe.Resize(orvyn.NewSize(width, 500))
	natural := lipgloss.Height(probe.layout.Render())

	if natural > probe.GetMinSize().Height {
		t.Logf("natural rendered height (%d) exceeds GetMinSize (%d), as expected: "+
			"GetMinSize measures the unwrapped description", natural, probe.GetMinSize().Height)
	}

	cases := []struct {
		name   string
		height int
	}{
		{"comfortably above natural size", natural + 10},
		{"just below natural size", natural - 1},
		{"far below natural size", 6},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := New()
			w.UpdateData(testItem())
			w.Resize(orvyn.NewSize(width, c.height))

			if got := lipgloss.Height(w.Render()); got != c.height {
				t.Errorf("rendered height = %d, want %d (the allocated height)",
					got, c.height)
			}
		})
	}
}

// TestRenderPadsWhenGivenMoreThanItNeeds guards the other half of the
// contract: MaxHeight must only clip, never shrink the box below an
// allocation it does not need. lipgloss's Height still has to do the padding.
func TestRenderPadsWhenGivenMoreThanItNeeds(t *testing.T) {
	const width = 60

	probe := New()
	probe.UpdateData(testItem())
	probe.Resize(orvyn.NewSize(width, 500))
	natural := lipgloss.Height(probe.layout.Render())

	allocated := natural + 20

	w := New()
	w.UpdateData(testItem())
	w.Resize(orvyn.NewSize(width, allocated))

	if got := lipgloss.Height(w.Render()); got != allocated {
		t.Errorf("rendered height = %d, want %d (the wider allocation, not the "+
			"smaller natural height %d)", got, allocated, natural)
	}
}
