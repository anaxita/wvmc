package responses

type SignIn struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewSignIn(accessToken, refreshToken string) SignIn {
	return SignIn{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
