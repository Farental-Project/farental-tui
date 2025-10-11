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

func LocationCreateBankAccount() *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/location/bank/create"
	r.SetError(api.ErrorResponse{})

	return r
}

func LocationBankGetFull() *resty.Request {
	r := client.R()

	r.Method = resty.MethodGet
	r.URL = "/location/bank/full"
	r.SetResult(api.InventoryResponse{})
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
