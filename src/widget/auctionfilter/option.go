package auctionfilter

// Option is one entry of a filter selector. Code is what the server expects,
// Label is what the player reads; an empty Code means the filter is off.
type Option struct {
	Code  string
	Label string
}

func (o Option) RenderValue() string {
	return o.Label
}
