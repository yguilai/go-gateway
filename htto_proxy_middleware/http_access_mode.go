package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/yguilai/go-gateway/dao"
	"github.com/yguilai/go-gateway/public"
)

// HTTPAccessModeMiddleware 匹配接入方式 基于请求信息
func HTTPAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := dao.ServiceManagerHandler.HTTPAccessMode(c)
		if err != nil {
			public.ResponseError(c, 1001, err)
			c.Abort()
			return
		}
		c.Set("service", service)
		c.Next()
	}
}
