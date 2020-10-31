package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yguilai/go-gateway/dao"
	"github.com/yguilai/go-gateway/public"
	"github.com/yguilai/go-gateway/reverse_proxy"
)

//匹配接入方式 基于请求信息
func HTTPReverseProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			public.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)

		lb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			public.ResponseError(c, 2002, err)
			c.Abort()
			return
		}
		trans, err := dao.TransportorHandler.GetTrans(serviceDetail)
		if err != nil {
			public.ResponseError(c, 2003, err)
			c.Abort()
			return
		}

		//创建 reverseproxy
		//使用 reverseproxy.ServerHTTP(c.Request,c.Response)
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(c, lb, trans)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
		return
	}
}
