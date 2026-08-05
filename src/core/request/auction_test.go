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
		"stat", "skill", "minStat", "outbid"} {
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

// Being outbid is a property of the auction, so it rides the browse endpoint
// as a filter rather than needing an endpoint of its own — but like every
// other filter it must be absent, not empty, when it is not asked for.
func TestAuctionGetAllSendsOutbidOnlyWhenSet(t *testing.T) {
	Init(resty.New())

	params := AuctionGetAll(1, api.AuctionFilter{OutbidOnly: true}).QueryParam

	if got := params.Get("outbid"); got != "true" {
		t.Errorf("outbid = %q, want %q", got, "true")
	}

	params = AuctionGetAll(1, api.AuctionFilter{}).QueryParam

	if _, ok := params["outbid"]; ok {
		t.Error("outbid sent for a filter that did not ask for it, want it omitted")
	}
}
