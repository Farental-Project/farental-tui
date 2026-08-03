package request

import (
	"farental/core/data/api"
	"testing"

	"github.com/go-resty/resty/v2"
)

// TestAuctionGetAllOmitsUnsetFilters covers what the server cares about: an
// unset filter must be absent from the query string, not present and empty.
// An empty stat= alongside an empty skill= is the contradiction /auction/all
// answers with a 400.
func TestAuctionGetAllOmitsUnsetFilters(t *testing.T) {
	Init(resty.New())

	params := AuctionGetAll(2, api.AuctionFilter{}).QueryParam

	if got := params.Get("page"); got != "2" {
		t.Errorf("page = %q, want %q", got, "2")
	}

	for _, name := range []string{"itemID", "kind", "slot", "weaponType",
		"stat", "skill", "minStat"} {
		if _, ok := params[name]; ok {
			t.Errorf("%s sent for an empty filter, want it omitted", name)
		}
	}
}

func TestAuctionGetAllSendsSetFilters(t *testing.T) {
	Init(resty.New())

	params := AuctionGetAll(1, api.AuctionFilter{
		ItemID:         12,
		Kind:           api.AuctionItemKindEquipment,
		SlotCode:       "hea",
		WeaponTypeCode: "sword",
		StatCode:       "str",
		MinStat:        -5,
		HasMinStat:     true,
	}).QueryParam

	want := map[string]string{
		"itemID":     "12",
		"kind":       "equipment",
		"slot":       "hea",
		"weaponType": "sword",
		"stat":       "str",
		"minStat":    "-5",
	}

	for name, value := range want {
		if got := params.Get(name); got != value {
			t.Errorf("%s = %q, want %q", name, got, value)
		}
	}
}

// A minimum with no stat or skill to refine narrows nothing server-side, so it
// is not worth sending. Zero is a legal minimum, which is why the filter
// carries HasMinStat rather than reading 0 as "unset".
func TestAuctionGetAllMinStatNeedsAStat(t *testing.T) {
	Init(resty.New())

	alone := AuctionGetAll(1, api.AuctionFilter{
		MinStat:    3,
		HasMinStat: true,
	}).QueryParam

	if _, ok := alone["minStat"]; ok {
		t.Error("minStat sent without a stat or skill, want it omitted")
	}

	withSkill := AuctionGetAll(1, api.AuctionFilter{
		SkillCode:  "swordsmanship",
		HasMinStat: true,
	}).QueryParam

	if got := withSkill.Get("minStat"); got != "0" {
		t.Errorf("minStat = %q, want %q", got, "0")
	}
}
