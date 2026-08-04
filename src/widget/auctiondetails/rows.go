package auctiondetails

import (
	"farental/art"
	"farental/core/data/api"
	"farental/widget/auctionlistitem"
	"fmt"
	"time"

	"github.com/halsten-dev/lokyn"
)

// row is one label/value line of the box.
type row struct {
	label string
	value string

	// highlight marks the money rows. They are rendered with the highlight
	// style so the price is the first thing read.
	highlight bool
}

// buildRows turns an auction into the lines the widget draws. It is kept apart
// from Render so the row set can be tested without a terminal.
//
// Quantity is deliberately absent: the dialog carries it in its title as
// "<item name> x<quantity>".
func buildRows(auction *api.AuctionResponse, now time.Time) []row {
	bid := fmt.Sprintf("%d%c", auction.CurrentBid, art.CharGrynars)

	if auctionlistitem.IsOwnBid(auction.CurrentBidderName) {
		bid = fmt.Sprintf("%s (%s)", bid, lokyn.L("you"))
	}

	rows := []row{
		{label: lokyn.L("Current bid"), value: bid, highlight: true},
	}

	if auction.DirectBuyPrice > 0 {
		rows = append(rows, row{
			label:     lokyn.L("Direct buy"),
			value:     fmt.Sprintf("%d%c", auction.DirectBuyPrice, art.CharGrynars),
			highlight: true,
		})
	}

	bidder := auction.CurrentBidderName

	if bidder == "" {
		bidder = lokyn.L("nobody")
	}

	return append(rows,
		row{label: lokyn.L("Current bidder"), value: bidder},
		row{label: lokyn.L("Seller"), value: auction.SellerName},
		row{
			label: lokyn.L("Ends in"),
			value: auctionlistitem.EndsIn(auction.EndTimestamp, now),
		},
	)
}
