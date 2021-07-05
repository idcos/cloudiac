package forms

type AccessTokenHandler struct {
	PageForm
	AccessToken string `json:"accessToken" form:"accessToken" binding:"required"`
}
