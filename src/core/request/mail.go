package request

import (
	"farental/core/data/api"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func MailSend(mail api.MailSendBody) *resty.Request {
	return post("/mail/sendBasic").SetBody(mail)
}

func MailSendWithAttachments(mail api.MailWithAttachmentsBody) *resty.Request {
	return post("/mail/sendWithAttachments").SetBody(mail)
}

func MailGetAll() *resty.Request {
	return get("/mail/all").SetResult([]api.MailBasicResponse{})
}

func MailGetOne(mailID uint) *resty.Request {
	return get("/mail/one").SetResult(api.MailBasicResponse{}).SetQueryParam("mailID", fmt.Sprint(mailID))
}

func MailGetAttachments(mailID uint) *resty.Request {
	return get("/mail/attachments").SetResult([]api.MailAttachmentResponse{}).SetQueryParam("mailID", fmt.Sprint(mailID))
}

func MailTransferAttachments(mailID uint) *resty.Request {
	return post("/mail/transferAttachments").SetBody(api.IDBody{ID: mailID})
}

func MailPay(mailID uint) *resty.Request {
	return post("/mail/pay").SetBody(api.IDBody{ID: mailID})
}

func MailSetRead(mailID uint, read bool) *resty.Request {
	return put("/mail/setRead").SetBody(api.MailSetReadBody{
		ID:     mailID,
		IsRead: read,
	})
}
