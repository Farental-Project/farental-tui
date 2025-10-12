package api

type BankAccountResponse struct {
	Inventory     InventoryResponse
	Rank          int
	MaxStackCount int
}
