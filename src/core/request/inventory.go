package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func InventoryGetFull() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/inventory/full"
	r.SetResult(api.InventoryResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func InventoryGetShareable() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/inventory/shareable"
	r.SetResult(api.InventoryResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func InventoryGetSellable() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/inventory/sellable"
	r.SetResult(api.InventoryResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func InventoryUseItem(itemID uint) *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/inventory/useItem"
	r.SetBody(api.IDBody{
		ID: itemID,
	})
	r.SetError(api.ErrorResponse{})

	return r
}

func InventoryEquipItem(itemID uint) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/inventory/equipItem"
	r.SetBody(api.IDBody{
		ID: itemID,
	})
	r.SetError(api.ErrorResponse{})

	return r
}

func InventoryGetEquippedItems() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/inventory/equippedItems"
	r.SetResult([]api.ItemResponse{})

	return r
}
