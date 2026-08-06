package request

import (
	"farental/core/data/api"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

// AuctionGetAll browses one page of the auction house, narrowed by filter.
//
// A filter left unset is left out of the query string rather than sent empty:
// the server reads a missing param as "no filter", and an empty stat= together
// with an empty skill= reads as the contradiction it rejects.
func AuctionGetAll(page int, filter api.AuctionFilter) *resty.Request {
	r := get("/auction/all").
		SetQueryParam("page", fmt.Sprintf("%d", page)).
		SetResult(api.AuctionListResponse{})

	if filter.ItemID != 0 {
		r.SetQueryParam("itemID", fmt.Sprintf("%d", filter.ItemID))
	}

	if filter.Kind != api.AuctionItemKindAny {
		r.SetQueryParam("kind", string(filter.Kind))
	}

	if filter.SlotCode != "" {
		r.SetQueryParam("slot", filter.SlotCode)
	}

	if filter.WeaponTypeCode != "" {
		r.SetQueryParam("weaponType", filter.WeaponTypeCode)
	}

	if filter.StatCode != "" {
		r.SetQueryParam("stat", filter.StatCode)
	}

	if filter.SkillCode != "" {
		r.SetQueryParam("skill", filter.SkillCode)
	}

	// A minimum with no stat to refine narrows nothing — the server drops it —
	// so it is only sent alongside the stat or skill it belongs to.
	if filter.HasMinStat && (filter.StatCode != "" || filter.SkillCode != "") {
		r.SetQueryParam("minStat", strconv.Itoa(filter.MinStat))
	}

	if filter.OutbidOnly {
		r.SetQueryParam("outbid", "true")
	}

	return r
}

// AuctionGetFilterOptions returns the filter vocabulary — kinds, equipment
// slots, weapon types, stats and skills — with localized labels, so the filter
// controls never hardcode codes.
func AuctionGetFilterOptions() *resty.Request {
	return get("/auction/filterOptions").
		SetResult(api.AuctionFilterOptionsResponse{})
}

func AuctionGetBids() *resty.Request {
	return get("/auction/bids").
		SetResult([]api.AuctionResponse{})
}

func AuctionGetOwn() *resty.Request {
	return get("/auction/own").
		SetResult([]api.AuctionResponse{})
}

func AuctionMakeBid(body api.AuctionBidBody) *resty.Request {
	return post("/auction/bid").
		SetBody(body)
}

func AuctionDirectBuy(body api.IDBody) *resty.Request {
	return post("/auction/buy").
		SetBody(body)
}

func AuctionCancel(body api.IDBody) *resty.Request {
	return post("/auction/cancel").
		SetBody(body)
}

func AuctionStart(body api.AuctionStartBody) *resty.Request {
	return post("/auction/start").SetBody(body)
}

func AuctionEstimate(body api.AuctionStartBody) *resty.Request {
	return post("/auction/estimate").
		SetBody(body).
		SetResult(api.AuctionPlanResponse{})
}
