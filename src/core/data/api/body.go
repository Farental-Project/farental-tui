package api

type IDBody struct {
	ID uint `validate:"required"`
}

type IDAmountBody struct {
	ID     uint `validate:"required"`
	Amount int  `validate:"required,number"`
}
