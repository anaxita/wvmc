package responses

type Refresh struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// NewRefresh ...
func NewRefresh(accessToken, refreshToken string) Refresh {
	return Refresh{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
