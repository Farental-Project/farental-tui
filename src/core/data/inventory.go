package data

import (
	"errors"
	"farental/core/data/api"
	"farental/internal/helper"
	"github.com/halsten-dev/lokyn"
	"sort"
)

type Inventory struct {
	MaxSize int
	Stacks  []api.StackResponse
}

func NewInventory(maxSize int) *Inventory {
	i := new(Inventory)

	i.MaxSize = maxSize
	i.Stacks = make([]api.StackResponse, 0)

	return i
}

// AddItem manage the addition of a certain amount of an Item in the inventory, managing stacks
// Can return a remaining count if the inventory is full.
func (inv *Inventory) AddItem(item *api.ItemResponse, count int) (int, error) {
	remainingCount, err := inv.addItem(item, count)

	if err != nil {
		return remainingCount, err
	}

	return remainingCount, nil
}

func (inv *Inventory) addItem(item *api.ItemResponse, count int) (int, error) {
	var itemStack api.StackResponse
	var freeSpace int
	var stackExists bool

	stackExists = false

	for i, stack := range inv.Stacks {
		if stack.ItemID != item.ID {
			continue
		} else {
			stackExists = true
		}

		if stack.Count == item.MaxStackCount {
			continue
		}

		freeSpace = item.MaxStackCount - stack.Count

		if count <= freeSpace {
			inv.Stacks[i].Count += count
			count = 0
		} else {
			count -= freeSpace
			inv.Stacks[i].Count = item.MaxStackCount
		}

		// No more item to add to the inventory, job is done.
		if count == 0 {
			return 0, nil
		}
	}

	// There are still objects to add to the inventory.
	for {
		if len(inv.Stacks) == inv.MaxSize {
			return count, errors.New(lokyn.L("Inventory is full"))
		}

		if item.IsUnique && stackExists {
			return count, errors.New(lokyn.L("Unique object stack is full"))
		}

		itemStack = api.StackResponse{
			ItemID: item.ID,
			Item:   *item,
		}

		if count <= item.MaxStackCount {
			itemStack.Count += count
			count = 0
		} else {
			count -= item.MaxStackCount
			itemStack.Count = item.MaxStackCount
		}

		inv.Stacks = append(inv.Stacks, itemStack)

		stackExists = true

		if count == 0 {
			break
		}
	}

	return 0, nil
}

// RemoveItem manage the removing of a certain amount of an Item in the inventory, managing stacks.
// Return the remaining count
func (inv *Inventory) RemoveItem(item *api.ItemResponse, count int) (int, error) {
	remainingCount, err := inv.removeItem(item, count)

	if err != nil {
		return remainingCount, err
	}

	return remainingCount, nil
}

func (inv *Inventory) removeItem(item *api.ItemResponse, count int) (int, error) {
	var inventoryStack *api.StackResponse
	var inventoryStackIndex int

	loopCounter := 0

	// Sort by stack count
	sort.Slice(inv.Stacks, func(i, j int) bool {
		return inv.Stacks[i].Count < inv.Stacks[j].Count
	})

	for {
		inventoryStack = nil

		for i, stack := range inv.Stacks {
			if stack.ItemID == item.ID {
				inventoryStack = &inv.Stacks[i]
				inventoryStackIndex = i
				break
			}
		}

		if inventoryStack == nil {
			if loopCounter == 0 {
				return count, errors.New(lokyn.L("Object does not exist in inventory"))
			} else {
				break
			}
		}

		if inventoryStack.Count > count {
			inventoryStack.Count = inventoryStack.Count - count
			count = 0
		} else {
			count -= inventoryStack.Count
			inv.Stacks = helper.SliceRemove(inv.Stacks, inventoryStackIndex)
		}

		if count == 0 {
			break
		}

		loopCounter++
	}

	return count, nil
}

func (inv *Inventory) RemoveIndex(index int) {
	if index < 0 || index >= len(inv.Stacks) {
		return
	}

	inv.Stacks = helper.SliceRemove(inv.Stacks, index)
}

func (inv *Inventory) ItemExists(item *api.ItemResponse) bool {
	for _, stack := range inv.Stacks {
		if stack.ItemID == item.ID {
			return true
		}
	}

	return false
}

func (inv *Inventory) ItemCountExists(itemID uint, amount int) bool {
	var count int

	count = 0

	for _, stack := range inv.Stacks {
		if stack.ItemID == itemID {
			if stack.Count >= amount {
				return true
			} else {
				count += stack.Count
			}

			if count >= amount {
				return true
			}
		}
	}

	return false
}
