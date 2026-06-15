package api

type IDBody struct {
	ID uint `validate:"required"`
}

type UUIDBody struct {
	ID []byte `validate:"required"`
}

type IDAmountBody struct {
	ID     uint `validate:"required"`
	Amount int  `validate:"required,number"`
}

type UUIDResponse struct {
	ID []byte
}
