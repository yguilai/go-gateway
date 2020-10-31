package http_proxy_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yguilai/go-gateway/dao"
	"github.com/yguilai/go-gateway/public"
)

func HTTPFlowLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			public.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)
		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			serviceLimiter, err := public.FlowLimiterHandler.GetLimiter(
				public.FlowServicePrefix+serviceDetail.Info.ServiceName,
				float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				public.ResponseError(c, 5001, err)
				c.Abort()
				return
			}
			if !serviceLimiter.Allow() {
				public.ResponseError(c, 5002, errors.New(fmt.Sprintf("service flow limit %v", serviceDetail.AccessControl.ServiceFlowLimit)))
				c.Abort()
				return
			}
		}

		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(
				public.FlowServicePrefix+serviceDetail.Info.ServiceName+"_"+c.ClientIP(),
				float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				public.ResponseError(c, 5003, err)
				c.Abort()
				return
			}
			if !clientLimiter.Allow() {
				public.ResponseError(c, 5002, errors.New(fmt.Sprintf("%v flow limit %v", c.ClientIP(), serviceDetail.AccessControl.ClientIPFlowLimit)))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
