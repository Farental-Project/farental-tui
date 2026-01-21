package api

type InventoryResponse struct {
	MaxStacks   int
	StacksCount int
	Stacks      []StackResponse
}

type StackResponse struct {
	ID     uint
	ItemID uint
	Item   ItemResponse
	Count  int
}
