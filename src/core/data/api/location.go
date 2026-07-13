package api

type LocationResponse struct {
	ID              uint
	Name            string
	Description     string
	LongDescription string

	Biome     LocationInfoResponse
	Type      LocationInfoResponse
	Continent LocationInfoResponse
	Features  []LocationFeatureResponse
}

func (l *LocationResponse) HaveFeature(code string) bool {
	if l.Features == nil {
		return false
	}

	for _, feature := range l.Features {
		if feature.Code == code {
			return true
		}
	}

	return false
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
