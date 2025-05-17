package api

type LocationResponse struct {
	ID          uint
	Name        string
	Description string

	Biome     LocationInfoResponse
	Type      LocationInfoResponse
	Continent LocationInfoResponse
	Features  []LocationFeatureResponse
}

type LocationInfoResponse struct {
	Code        string
	Name        string
	Description string
}

type LocationFeatureResponse struct {
	LocationInfoResponse
	IsAction bool
}
