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
	rows := []row{
		{label: lokyn.L("Current bid"), value: fmt.Sprintf("%d%c", auction.CurrentBid, art.CharGrynars), highlight: true},
	}

	if auction.DirectBuyPrice > 0 {
		rows = append(rows, row{
			label:     lokyn.L("Direct buy"),
			value:     fmt.Sprintf("%d%c", auction.DirectBuyPrice, art.CharGrynars),
			highlight: true,
		})
	}

	bidder := auction.CurrentBidderName

	switch {
	case bidder == "":
		bidder = lokyn.L("nobody")

	case auctionlistitem.IsOwnBid(bidder):
		bidder = lokyn.L("you")
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
