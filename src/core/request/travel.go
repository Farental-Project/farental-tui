package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func TravelGetAvailable() *resty.Request {
	return get("/travel/available").SetResult([]api.TravelResponse{})
}

func TravelRelayGetAvailable() *resty.Request {
	return get("/travel/relay/available").SetResult([]api.TravelRelayResponse{})
}

func TravelStart(travelID uint) *resty.Request {
	return post("/travel/start").SetBody(api.TravelStartBody{
		TravelID: travelID,
	})
}

func TravelRelayStart(destLocationID uint) *resty.Request {
	return post("/travel/relay/start").SetBody(api.TravelRelayStartBody{
		DestLocationID: destLocationID,
	})
}
