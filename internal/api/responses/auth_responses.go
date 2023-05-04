package responses

import (
	"github.com/anaxita/wvmc/internal/entity"
)

type PostSignIn struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	Role         entity.Role `json:"role"`
}

func NewPostSignIn(data entity.SignIn) PostSignIn {
	return PostSignIn{
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		Role:         data.Role,
	}
}

type PostRefresh struct {
	PostSignIn
}

func NewPostRefresh(data entity.SignIn) PostRefresh {
	return PostRefresh{
		PostSignIn: NewPostSignIn(data),
	}
}
