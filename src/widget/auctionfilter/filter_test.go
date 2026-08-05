package auctionfilter

import (
	"farental/core/data/api"
	"testing"
)

func TestBuildFilterEmpty(t *testing.T) {
	got := buildFilter("", "", "", "", "", "", "")

	if got != (api.AuctionFilter{}) {
		t.Errorf("buildFilter with no values = %+v, want the zero filter", got)
	}
}

func TestBuildFilterAllSet(t *testing.T) {
	got := buildFilter("", "equipment", "hea", "sword", "str", "", "5")

	want := api.AuctionFilter{
		Kind:           api.AuctionItemKindEquipment,
		SlotCode:       "hea",
		WeaponTypeCode: "sword",
		StatCode:       "str",
		MinStat:        5,
		HasMinStat:     true,
	}

	if got != want {
		t.Errorf("buildFilter = %+v, want %+v", got, want)
	}
}

// The server ignores a minimum with no stat to refine, so the widget must not
// claim one is set.
func TestBuildFilterMinStatNeedsAStat(t *testing.T) {
	got := buildFilter("", "", "", "", "", "", "5")

	if got.HasMinStat {
		t.Errorf("HasMinStat = true with no stat or skill, want false")
	}
}

func TestBuildFilterNegativeMinStat(t *testing.T) {
	got := buildFilter("", "", "", "", "", "swordsmanship", "-5")

	if got.SkillCode != "swordsmanship" {
		t.Errorf("SkillCode = %q, want %q", got.SkillCode, "swordsmanship")
	}

	if !got.HasMinStat || got.MinStat != -5 {
		t.Errorf("MinStat = %d (has: %v), want -5 (true)", got.MinStat, got.HasMinStat)
	}
}

func TestBuildFilterUnparsableMinStat(t *testing.T) {
	got := buildFilter("", "", "", "", "str", "", "-")

	if got.HasMinStat {
		t.Errorf("HasMinStat = true for an unparsable minimum, want false")
	}
}

func TestBuildFilterScope(t *testing.T) {
	cases := []struct {
		name  string
		scope string
		want  bool
	}{
		{"empty scope browses the whole house", "", false},
		{"outbid scope narrows to the player's lost leads", "outbid", true},
	}

	for _, tc := range cases {
		f := buildFilter(tc.scope, "", "", "", "", "", "")

		if f.OutbidOnly != tc.want {
			t.Errorf("%s: OutbidOnly = %v, want %v", tc.name, f.OutbidOnly, tc.want)
		}
	}
}

// The scope narrows which auctions are searched; the rest narrow within them.
// Nothing about the two is exclusive, so both have to survive one call.
func TestBuildFilterScopeCombinesWithItemFilters(t *testing.T) {
	f := buildFilter("outbid", "equipment", "hea", "sword", "str", "", "5")

	if !f.OutbidOnly {
		t.Error("OutbidOnly = false, want true")
	}

	if f.Kind != api.AuctionItemKindEquipment {
		t.Errorf("Kind = %q, want %q", f.Kind, api.AuctionItemKindEquipment)
	}

	if f.SlotCode != "hea" {
		t.Errorf("SlotCode = %q, want %q", f.SlotCode, "hea")
	}

	if !f.HasMinStat || f.MinStat != 5 {
		t.Errorf("MinStat = %d (set %v), want 5 (set true)", f.MinStat, f.HasMinStat)
	}
}
