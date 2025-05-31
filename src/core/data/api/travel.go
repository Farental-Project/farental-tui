package api

type TravelStartBody struct {
	TravelID uint `validate:"required"`
}

type TravelResponse struct {
	ID                       uint
	FromLocation             LocationResponse
	ToLocation               LocationResponse
	DestLocation             LocationResponse
	Duration                 float64
	Price                    int
	RequestedLocationFeature LocationFeatureResponse
}
