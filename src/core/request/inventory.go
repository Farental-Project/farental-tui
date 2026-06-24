package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func InventoryGetFull() *resty.Request {
	return get("/inventory/full").SetResult(api.InventoryResponse{})
}

func InventoryGetShareable() *resty.Request {
	return get("/inventory/shareable").SetResult(api.InventoryResponse{})
}

func InventoryGetSellable() *resty.Request {
	return get("/inventory/sellable").SetResult(api.InventoryResponse{})
}

func InventoryUseItem(itemID uint) *resty.Request {
	return post("/inventory/useItem").SetBody(api.IDBody{ID: itemID})
}

func InventoryEquipItem(itemID uint) *resty.Request {
	return post("/inventory/equipItem").SetBody(api.IDBody{ID: itemID})
}

func InventoryUnequipItem(itemID uint) *resty.Request {
	return post("/inventory/unequipItem").SetBody(api.IDBody{ID: itemID})
}

func InventoryGetEquippedItems() *resty.Request {
	return get("/inventory/equippedItems").SetResult([]api.ItemResponse{})
}
