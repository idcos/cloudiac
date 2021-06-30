package forms

type AccessTokenHandler struct {
	BaseForm
	AccessToken string `json:"accessToken" form:"accessToken" binding:"required"`
}
