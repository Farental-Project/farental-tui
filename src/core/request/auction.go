package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func AuctionGetAll(page int, itemID uint) *resty.Request {
	return get("/auction/all").
		SetQueryParam("page", fmt.Sprintf("%d", page)).
		SetQueryParam("itemID", fmt.Sprintf("%d", itemID)).
		SetResult(api.AuctionListResponse{})
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
