package requests

type PostRefresh struct {
	RefreshToken string `json:"refresh_token"`
}

type PostSignIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
