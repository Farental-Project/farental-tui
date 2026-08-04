package auctiondetails

import (
	"farental/core/data/api"
	"farental/internal/context"
	"os"
	"testing"
	"time"

	"github.com/halsten-dev/lokyn"
)

func TestMain(m *testing.M) {
	lokyn.Init()
	lokyn.SetLanguage("en")

	os.Exit(m.Run())
}

// labels pulls the label column out of the rows so a test can assert the row
// set without caring about the values.
func labels(rows []row) []string {
	out := make([]string, 0, len(rows))

	for _, r := range rows {
		out = append(out, r.label)
	}

	return out
}

func find(t *testing.T, rows []row, label string) row {
	t.Helper()

	for _, r := range rows {
		if r.label == label {
			return r
		}
	}

	t.Fatalf("no row labelled %q in %v", label, labels(rows))

	return row{}
}

func TestBuildRowsWithoutDirectBuy(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	auction := api.AuctionResponse{
		CurrentBid:        120,
		DirectBuyPrice:    0,
		CurrentBidderName: "Jane Roe",
		SellerName:        "John Doe",
		EndTimestamp:      now.Add(2 * time.Hour),
	}

	rows := buildRows(&auction, now)

	want := []string{"Current bid", "Current bidder", "Seller", "Ends in"}

	if got := labels(rows); len(got) != len(want) {
		t.Fatalf("labels = %v, want %v", got, want)
	}

	for i, l := range labels(rows) {
		if l != want[i] {
			t.Errorf("row %d label = %q, want %q", i, l, want[i])
		}
	}
}

func TestBuildRowsWithDirectBuy(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	auction := api.AuctionResponse{
		CurrentBid:        120,
		DirectBuyPrice:    500,
		CurrentBidderName: "Jane Roe",
		SellerName:        "John Doe",
		EndTimestamp:      now.Add(2 * time.Hour),
	}

	rows := buildRows(&auction, now)

	directBuy := find(t, rows, "Direct buy")

	if directBuy.value != "500Ǥ" {
		t.Errorf("Direct buy value = %q, want %q", directBuy.value, "500Ǥ")
	}

	if !directBuy.highlight {
		t.Error("Direct buy should be highlighted as a money row")
	}
}

func TestBuildRowsEmptyBidderReadsNobody(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	auction := api.AuctionResponse{
		CurrentBid:   0,
		SellerName:   "John Doe",
		EndTimestamp: now.Add(2 * time.Hour),
	}

	rows := buildRows(&auction, now)

	if got := find(t, rows, "Current bidder").value; got != "nobody" {
		t.Errorf("Current bidder value = %q, want %q", got, "nobody")
	}
}

func TestBuildRowsMarksOwnBid(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	original := context.CharacterInfo
	defer func() { context.CharacterInfo = original }()

	context.CharacterInfo = &api.CharacterInfoResponse{
		FirstName: "John",
		LastName:  "Doe",
	}

	auction := api.AuctionResponse{
		CurrentBid:        120,
		CurrentBidderName: "John Doe",
		SellerName:        "Jane Roe",
		EndTimestamp:      now.Add(2 * time.Hour),
	}

	rows := buildRows(&auction, now)

	if got := find(t, rows, "Current bid").value; got != "120Ǥ (you)" {
		t.Errorf("Current bid value = %q, want %q", got, "120Ǥ (you)")
	}
}

func TestBuildRowsEndsInUsesSharedFormatter(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	auction := api.AuctionResponse{
		SellerName:   "John Doe",
		EndTimestamp: now.Add(4*time.Hour + 2*time.Minute),
	}

	rows := buildRows(&auction, now)

	if got := find(t, rows, "Ends in").value; got != "4h02" {
		t.Errorf("Ends in value = %q, want %q", got, "4h02")
	}
}
