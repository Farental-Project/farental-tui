package api

type CurrencyCode string

const (
	Grynars CurrencyCode = "grynars"
)

type CurrencyResponse struct {
	Code   CurrencyCode
	Amount int
}
