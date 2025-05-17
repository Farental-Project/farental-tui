package api

type ErrorElement struct {
	Message string
}

type ErrorResponse struct {
	Errors []ErrorElement
}

type DataResponse[T any] struct {
	Data T
}
