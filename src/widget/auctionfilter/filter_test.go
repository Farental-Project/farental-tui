package auctionfilter

import (
	"farental/core/data/api"
	"testing"
)

func TestBuildFilterEmpty(t *testing.T) {
	got := buildFilter("", "", "", "", "", "")

	if got != (api.AuctionFilter{}) {
		t.Errorf("buildFilter with no values = %+v, want the zero filter", got)
	}
}

func TestBuildFilterAllSet(t *testing.T) {
	got := buildFilter("equipment", "hea", "sword", "str", "", "5")

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
	got := buildFilter("", "", "", "", "", "5")

	if got.HasMinStat {
		t.Errorf("HasMinStat = true with no stat or skill, want false")
	}
}

func TestBuildFilterNegativeMinStat(t *testing.T) {
	got := buildFilter("", "", "", "", "swordsmanship", "-5")

	if got.SkillCode != "swordsmanship" {
		t.Errorf("SkillCode = %q, want %q", got.SkillCode, "swordsmanship")
	}

	if !got.HasMinStat || got.MinStat != -5 {
		t.Errorf("MinStat = %d (has: %v), want -5 (true)", got.MinStat, got.HasMinStat)
	}
}

func TestBuildFilterUnparsableMinStat(t *testing.T) {
	got := buildFilter("", "", "", "str", "", "-")

	if got.HasMinStat {
		t.Errorf("HasMinStat = true for an unparsable minimum, want false")
	}
}
