package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func Login() *resty.Request {
	return post("/auth/login").SetResult(api.AuthSuccessResponse{})
}

func SignUp(username, email, password, confirmPassword, lang string) *resty.Request {
	return post("/auth/register").
		SetBody(api.AuthSignUpBody{
			Username:        username,
			Email:           email,
			Password:        password,
			PasswordConfirm: confirmPassword,
		}).
		SetQueryParam("lang", lang)
}

func AuthInfo() *resty.Request {
	return get("/auth/info").SetResult(api.UserResponse{})
}

func AuthSetSettings(body api.UserSettingsBody) *resty.Request {
	return post("/auth/setSettings").SetBody(body)
}

func AuthSendFeedback(body api.SendFeedbackBody) *resty.Request {
	return post("/auth/sendFeedback").SetBody(body)
}
