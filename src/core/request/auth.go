package request

import (
	"farental/core/data/api"

	"github.com/go-resty/resty/v2"
)

func Login() *resty.Request {
	r := client.R()
	r.Method = resty.MethodPost
	r.URL = "/auth/login"
	r.SetResult(api.AuthSuccessResponse{})
	r.SetError(api.ErrorResponse{})

	return r
}

func SignUp(username, email, password, confirmPassword string) *resty.Request {
	r := client.R()

	r.Method = resty.MethodPost
	r.URL = "/auth/register"
	r.SetBody(api.AuthSignUpBody{
		Username:        username,
		Email:           email,
		Password:        password,
		PasswordConfirm: confirmPassword,
	})
	r.SetError(api.ErrorResponse{})

	return r
}

func AuthInfo() *resty.Request {
	r := client.R()
	r.Method = resty.MethodGet
	r.URL = "/auth/info"
	r.SetResult(api.DataResponse[api.UserResponse]{})

	return r
}
