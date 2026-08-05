package auctionlistitem

import (
	"farental/core/data/api"
	"farental/internal/context"
	"testing"
)

func TestIsOwnBid(t *testing.T) {
	cases := []struct {
		name   string
		info   *api.CharacterInfoResponse
		bidder string
		want   bool
	}{
		{
			name:   "no character loaded",
			info:   nil,
			bidder: "John Doe",
			want:   false,
		},
		{
			name:   "no bidder yet",
			info:   &api.CharacterInfoResponse{FirstName: "John", LastName: "Doe"},
			bidder: "",
			want:   false,
		},
		{
			name:   "someone else holds the bid",
			info:   &api.CharacterInfoResponse{FirstName: "John", LastName: "Doe"},
			bidder: "Jane Roe",
			want:   false,
		},
		{
			name:   "player holds the bid",
			info:   &api.CharacterInfoResponse{FirstName: "John", LastName: "Doe"},
			bidder: "John Doe",
			want:   true,
		},
		{
			name:   "partial name is not a match",
			info:   &api.CharacterInfoResponse{FirstName: "John", LastName: "Doe"},
			bidder: "John",
			want:   false,
		},
	}

	original := context.CharacterInfo
	defer func() { context.CharacterInfo = original }()

	for _, c := range cases {
		context.CharacterInfo = c.info

		if got := IsOwnBid(c.bidder); got != c.want {
			t.Errorf("%s: IsOwnBid(%q) = %v, want %v",
				c.name, c.bidder, got, c.want)
		}
	}
}
