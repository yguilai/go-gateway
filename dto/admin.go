package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/yguilai/go-gateway/public"
	"time"
)

type AdminInput struct {
	Username string `json:"username" form:"username" comment:"用户名" example:"admin" validate:"required"`
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`
}

type AdminOutput struct {
	Token string `json:"token" comment:"jwt token" example:"token" `
}

type AdminSessionInfo struct {
	Id           int       `json:"id"`
	Username     string    `json:"username"`
	LoginTime    time.Time `json:"login_time"`
}

type AdminInfoOutput struct {
	AdminSessionInfo
	// ...其他字段
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}



func (t *AdminInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, t)
}

type UpdatePwdInput struct {
	Password string `json:"password"`
}

func (t *UpdatePwdInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, t)
}