package entity

type SignIn struct {
	AccessToken  string
	RefreshToken string
	Role         UserRole
}
