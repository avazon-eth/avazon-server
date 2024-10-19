package dto

type GoogleCode struct {
	Code string `json:"code" binding:"required"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
