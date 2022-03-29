package configService

import (
	"github.com/CHainGate/backend/configApi"
	"testing"
)

func TestAuthenticationApiService_Login(t *testing.T) {
	service := AuthenticationApiService{}
	loginRequestDto := configApi.LoginRequestDto{}

	login, err := service.Login(nil, loginRequestDto)
	if err != nil {
		return
	}

	t.Log(login)
}
