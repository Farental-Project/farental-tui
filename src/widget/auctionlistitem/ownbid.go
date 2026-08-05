package auctionlistitem

import (
	"farental/internal/context"
	"fmt"
)

// IsOwnBid reports whether bidderName is the logged-in character. The server
// sends readable names, not IDs, so comparing the full name is the only way to
// tell the player they already hold the bid.
func IsOwnBid(bidderName string) bool {
	info := context.CharacterInfo

	if info == nil || bidderName == "" {
		return false
	}

	return bidderName == fmt.Sprintf("%s %s", info.FirstName, info.LastName)
}
