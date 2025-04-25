package domain

type SignInRequest struct {
	CodeVerifier string `json:"code_verifier"`
	RedirectUrl  string `json:"redirect_uri"`
	Code         string `json:"code"`
	DeviceId     string `json:"device_id"`
	State        string `json:"state"`
}

type SignInData struct {
	User         UserHeader
	AccessToken  string
	RefreshToken string
}
