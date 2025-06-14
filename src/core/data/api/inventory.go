package api

type InventoryResponse struct {
	Stacks []StackResponse
}

type StackResponse struct {
	ID     uint
	ItemID uint
	Item   ItemResponse
	Count  int
}
