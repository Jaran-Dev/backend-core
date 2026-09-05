package http

type OAuthCallbackQuery struct {
	Code  string `form:"code" binding:"required"`
	State string `form:"state" binding:"required"`
}

type OAuthLinkQuery struct {
	Code  string `form:"code" binding:"required"`
	State string `form:"state" binding:"required"`
}
