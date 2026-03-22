package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func LocationGetCharacters() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/location/characters"
	r.SetResult([]api.CharacterBasicWithActivityResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationTavernSleep() *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/location/tavern/sleep"
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationTavernRegen() *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/location/tavern/regen"
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationMerchantBuyItem(itemID uint, amount int) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/location/merchant/buyItem"
	r.SetBody(api.IDAmountBody{
		ID:     itemID,
		Amount: amount,
	})
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationMerchantGetBuyableItem() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/location/merchant/buyableItems"
	r.SetResult([]api.ItemResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationMerchantSellItem(itemID uint, amount int) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "location/merchant/sellItem"
	r.SetBody(api.IDAmountBody{
		ID:     itemID,
		Amount: amount,
	})
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationCreateBankAccount() *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/location/bank/create"
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationBankGetAccount() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/location/bank/account"
	r.SetResult(api.BankAccountResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationBankUpgradeAccount() *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/location/bank/buyRankUpgrade"
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationBankTransferTo(itemID uint, amount int) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/location/bank/transferTo"
	r.SetBody(api.IDAmountBody{
		ID:     itemID,
		Amount: amount,
	})
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationBankTransferFrom(itemID uint, amount int) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/location/bank/transferFrom"
	r.SetBody(api.IDAmountBody{
		ID:     itemID,
		Amount: amount,
	})
	r.SetError(api.ErrorResponse{})

	return r
}
