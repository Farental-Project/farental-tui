package auctionfilter

import (
	"farental/core/data/api"
	"strconv"
)

// buildFilter assembles the query from raw control values. Kept free of the
// widgets so the rules the server enforces can be tested directly.
func buildFilter(kind, slot, weaponType, statCode, skillCode, minStat string) api.AuctionFilter {
	f := api.AuctionFilter{
		Kind:           api.AuctionItemKind(kind),
		SlotCode:       slot,
		WeaponTypeCode: weaponType,
		StatCode:       statCode,
		SkillCode:      skillCode,
	}

	// A minimum narrows a stat; alone it is a no-op the server drops.
	if statCode == "" && skillCode == "" {
		return f
	}

	value, err := strconv.Atoi(minStat)

	if err != nil {
		return f
	}

	f.MinStat = value
	f.HasMinStat = true

	return f
}
