package middleware

import (
	"errors"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/yguilai/go-gateway/public"
)

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if adminInfo := session.Get(public.AdminSessionInfoKey); adminInfo == nil {
			public.ResponseError(c, public.InternalErrorCode, errors.New("admin not login"))
			c.Abort()
			return
		}
		c.Next()
	}
}
