package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func LocationGetCharacters() *resty.Request {
	return get("/location/characters").SetResult([]api.CharacterBasicWithActivityResponse{})
}

func LocationTavernSleep() *resty.Request {
	return post("/location/tavern/sleep")
}

func LocationTavernRegen() *resty.Request {
	return post("/location/tavern/regen")
}

func LocationMerchantBuyItem(itemID uint, amount int) *resty.Request {
	return post("/location/merchant/buyItem").SetBody(api.IDAmountBody{
		ID:     itemID,
		Amount: amount,
	})
}

func LocationMerchantGetBuyableItem() *resty.Request {
	return get("/location/merchant/buyableItems").SetResult([]api.ItemResponse{})
}

func LocationMerchantSellItem(itemID uint, amount int) *resty.Request {
	return post("/location/merchant/sellItem").SetBody(api.IDAmountBody{
		ID:     itemID,
		Amount: amount,
	})
}

func LocationCreateBankAccount() *resty.Request {
	return post("/location/bank/create")
}

func LocationBankGetAccount() *resty.Request {
	return get("/location/bank/account").SetResult(api.BankAccountResponse{})
}

func LocationBankUpgradeAccount() *resty.Request {
	return post("/location/bank/buyRankUpgrade")
}

func LocationBankTransferTo(itemID uint, amount int) *resty.Request {
	return post("/location/bank/transferTo").SetBody(api.IDAmountBody{
		ID:     itemID,
		Amount: amount,
	})
}

func LocationBankTransferFrom(itemID uint, amount int) *resty.Request {
	return post("/location/bank/transferFrom").SetBody(api.IDAmountBody{
		ID:     itemID,
		Amount: amount,
	})
}
