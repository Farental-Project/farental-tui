package api

import (
	"time"

	"github.com/halsten-dev/lokyn"
)

type AuctionDuration int

const (
	// 48h
	AuctionDurationShort AuctionDuration = 48

	// 72h
	AuctionDurationLong AuctionDuration = 72

	// 96h
	AuctionDurationVeryLong AuctionDuration = 96
)

func (a AuctionDuration) IsValid() bool {
	switch a {
	case AuctionDurationShort, AuctionDurationLong, AuctionDurationVeryLong:
		return true
	}

	return false
}

func (a AuctionDuration) RenderValue() string {
	switch a {
	case 48:
		return lokyn.L("Short auction (48h)")
	case 72:
		return lokyn.L("Long auction (72h)")
	case 96:
		return lokyn.L("Very long auction (96h)")
	}

	return ""
}

// AuctionStatus records how an auction ended, which EndTimestamp alone cannot
// express: sold, cancelled and expired-unsold all used to be encoded as
// "end_timestamp is in the past", leaving the settlement engine unable to tell
// which had already been handled inline or which mail an outcome deserved.
type AuctionStatus int

const (
	// AuctionStatusActive is running and biddable until EndTimestamp.
	AuctionStatusActive AuctionStatus = iota

	// AuctionStatusSold was bought directly, or won by a bidder at expiry.
	AuctionStatusSold

	// AuctionStatusExpired ran out with no bidder; items go back to the seller.
	AuctionStatusExpired

	// AuctionStatusCancelled was pulled by the seller before expiry.
	AuctionStatusCancelled
)

func (s AuctionStatus) IsValid() bool {
	switch s {
	case AuctionStatusActive, AuctionStatusSold, AuctionStatusExpired, AuctionStatusCancelled:
		return true
	}

	return false
}

type AuctionStartBody struct {
	ItemID uint `validate:"required"`

	TotalQuantity int `validate:"required,min=1"`
	StackSize     int `validate:"required,min=1"`

	UnitStartBid       int `validate:"required,min=1"`
	UnitDirectBuyPrice int `validate:"min=0"`

	Duration AuctionDuration `validate:"required"`
}

// AuctionPlanResponse tells the client what a listing request produces, so it
// can show how many auctions will be created and what they cost before the
// seller commits — and so a request cut short by the slot limit says so instead
// of silently listing fewer items than asked.
//
// Returned by both /auction/estimate (nothing created) and /auction/start.
type AuctionPlanResponse struct {
	// RequestedAuctions is how many the quantity and stack size imply.
	RequestedAuctions int

	// Auctions is how many will be — or were — created, after the slot limit.
	Auctions int

	// Truncated is true when the slot limit cut the request short.
	Truncated bool

	// QuantityListed is the total number of items across those auctions.
	QuantityListed int

	// TotalTax is the listing tax in grynars.
	TotalTax int

	// SlotsRemaining is how many auction slots the seller had free.
	SlotsRemaining int
}

// AuctionListResponse is one page of the auction house.
//
// Browsing is paged because each entry carries a fully preloaded item; Total
// lets a client show how many more there are without fetching them.
type AuctionListResponse struct {
	Auctions []AuctionResponse
	Total    int64
	Page     int
	PageSize int
}

// AuctionItemKind is the bucket filter /auction/all accepts. The codes mirror
// the server's repository.AuctionItemKind; the readable labels come from
// /auction/filterOptions, which returns them in the player's language.
type AuctionItemKind string

const (
	AuctionItemKindAny        AuctionItemKind = ""
	AuctionItemKindConsumable AuctionItemKind = "consumable"
	AuctionItemKindEquipment  AuctionItemKind = "equipment"
	AuctionItemKindMaterial   AuctionItemKind = "material"
)

func (k AuctionItemKind) IsValid() bool {
	switch k {
	case AuctionItemKindAny, AuctionItemKindConsumable,
		AuctionItemKindEquipment, AuctionItemKindMaterial:
		return true
	}

	return false
}

// AuctionFilterKindResponse is one value of the kind filter, with a label to
// show. Code is what /auction/all expects back.
type AuctionFilterKindResponse struct {
	Code string
	Name string
}

// AuctionFilterOptionsResponse is the vocabulary behind the auction house's
// filter controls, returned by /auction/filterOptions. Building the controls
// from it rather than from hardcoded lists keeps the labels localized and the
// codes in step with the server.
type AuctionFilterOptionsResponse struct {
	Kinds          []AuctionFilterKindResponse
	EquipmentSlots []EquipmentSlotResponse
	WeaponTypes    []WeaponTypeResponse
	Stats          []StatResponse
	Skills         []SkillResponse
}

// AuctionFilter is one browse query. Every field is optional; the zero value
// browses the whole house.
//
// Codes, not IDs: the server resolves them and answers an unknown one with a
// readable rejection, so nothing here needs to know about database identifiers.
// The seller's own listings are excluded by the server, not asked for here.
type AuctionFilter struct {
	// ItemID narrows to a single item when non-zero.
	ItemID uint

	Kind AuctionItemKind

	SlotCode       string
	WeaponTypeCode string

	// StatCode and SkillCode both name a stat — a skill stands for its
	// primordial stat — so the server rejects a request carrying both. The
	// filter controls have to offer them as a single choice.
	StatCode  string
	SkillCode string

	// MinStat applies only when HasMinStat is set. Stat values are signed, so
	// a defaulted minimum of 0 would quietly hide every item carrying a
	// penalty.
	MinStat    int
	HasMinStat bool
}

type AuctionBidBody struct {
	ID  uint `validate:"required"`
	Bid int  `validate:"required,min=1"`
}

type AuctionResponse struct {
	ID uint

	Item     ItemResponse
	Quantity int

	CurrentBid     int
	DirectBuyPrice int

	CurrentBidderName string

	SellerName string

	Duration AuctionDuration

	EndTimestamp time.Time
}
