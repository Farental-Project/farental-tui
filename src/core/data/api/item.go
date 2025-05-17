package api

type ItemResponse struct {
	ID             uint
	Name           string
	Description    string
	IsUnique       bool
	IsUsable       bool
	MaxStackCount  int
	EquipmentSlot  *EquipmentSlotResponse
	EquipmentStats *[]EquipmentStatResponse
}

type EquipmentSlotResponse struct {
	Code string
	Name string
}

type EquipmentStatResponse struct {
	Stat  StatResponse
	Value int
}
