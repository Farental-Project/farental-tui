package api

import (
	"time"
)

type MailSendBody struct {
	ReceiverCharacterID uint   `validate:"required"`
	Subject             string `validate:"required"`
	Content             string `validate:"required"`
}

type MailWithAttachmentsBody struct {
	MailSendBody
	IsAgainstPayment bool
	MoneyAmount      int `validate:"min=0"`
	Items            []MailAttachment
}

type MailAttachment struct {
	ItemID uint `validate:"required"`
	Amount int
}

type MailSetReadBody struct {
	ID     uint `validate:"required"`
	IsRead bool
}

type MailBasicResponse struct {
	ID               uint
	DeliveredAt      time.Time
	SenderName       string
	Subject          string
	Content          string
	IsAgainstPayment bool
	IsRead           bool
	MoneyAmount      int
	HaveAttachments  bool
}
