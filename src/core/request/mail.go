package request

import (
	"farental/core/data/api"
	"fmt"
	"github.com/go-resty/resty/v2"
)

func MailGetAll() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/mail/all"
	r.SetResult([]api.MailBasicResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func MailGetOne(mailID uint) *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/mail/one"
	r.SetQueryParam("mailID", fmt.Sprint(mailID))
	r.SetResult(api.MailBasicResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func MailGetAttachments(mailID uint) *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/mail/attachments"
	r.SetQueryParam("mailID", fmt.Sprint(mailID))
	r.SetResult([]api.MailAttachmentResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func MailTransferAttachments(mailID uint) *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/mail/transferAttachments"
	r.SetBody(api.IDBody{ID: mailID})
	r.SetError(api.ErrorResponse{})

	return r
}

func MailPay(mailID uint) *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/mail/pay"
	r.SetBody(api.IDBody{ID: mailID})
	r.SetError(api.ErrorResponse{})

	return r
}
