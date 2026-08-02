package api

import "slices"

type InventoryResponse struct {
	MaxStacks   int
	StacksCount int
	Stacks      []StackResponse
}

func (i *InventoryResponse) CreateGroupedStackResponseList() []StackResponse {
	items := make([]StackResponse, 0)

	for _, s := range i.Stacks {
		index := slices.IndexFunc(items,
			func(stack StackResponse) bool {
				return stack.ItemID == s.ItemID
			})

		// Non-existing index
		if index == -1 {
			listItem := s

			items = append(items, listItem)
			continue
		}

		items[index].Count += s.Count
	}

	return items
}

type StackResponse struct {
	ID     uint
	ItemID uint
	Item   ItemResponse
	Count  int
}
