package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yguilai/go-gateway/common/lib"
	"github.com/yguilai/go-gateway/dao"
	"github.com/yguilai/go-gateway/dto"
	"github.com/yguilai/go-gateway/public"
)

func RegisterService(r *gin.RouterGroup) {
	r.GET("", ServiceList)
	r.DELETE("/:id", ServiceDelete)
	r.GET("/:id", ServiceDetail)
	r.GET("/:id/stat", ServiceStat)

	r.POST("/http", ServiceAddHTTP)
	r.PUT("/http", ServiceUpdateHTTP)
	r.POST("/tcp", ServiceAddTCP)
	r.PUT("/tcp", ServiceUpdateTCP)
	r.POST("/grpc", ServiceAddGRPC)
	r.PUT("/grpc", ServiceUpdateGRPC)
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /services/list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_no query int true "当前页数"
// @Success 200 {object} public.Response{data=dto.ServiceListOutput} "success"
// @Router /services [GET]
func ServiceList(c *gin.Context) {
	p := &dto.ServiceListInput{}
	if err := p.BindValidParam(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	tx, err := lib.GetGormPool(DB_SCOPE)
	if err != nil {
		public.ResponseError(c, public.GetGormPoolErrorCode, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	infos, total, err := serviceInfo.PageList(c, tx, p)
	if err != nil {
		public.ResponseError(c, 2002, err)
		return
	}

	list := make([]dto.ServiceListItemOutput, 0)

	for _, info := range infos {
		detail, err := info.ServiceDetail(c, tx, &info)
		if err != nil {
			public.ResponseError(c, 2003, err)
			return
		}

		serviceAddr := "unknow"
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		switch detail.Info.LoadType {
		case public.LoadTypeHTTP:
			if detail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL && detail.HTTPRule.NeedHttps == 1 {
				serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, detail.HTTPRule.Rule)
			} else if detail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL && detail.HTTPRule.NeedHttps == 0 {
				serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, detail.HTTPRule.Rule)
			} else if detail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
				serviceAddr = fmt.Sprintf("%s:%d", clusterIP, detail.TCPRule.Port)
			}
		case public.LoadTypeTCP:
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, detail.TCPRule.Port)
		case public.LoadTypeGRPC:
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, detail.GRPCRule.Port)
		}

		ipList := detail.LoadBalance.GetIPListByModel()
		counter, err := public.FlowCounterHandler.GetCounter(public.FlowServicePrefix + info.ServiceName)
		if err != nil {
			public.ResponseError(c, 2004, err)
			return
		}

		item := dto.ServiceListItemOutput{
			ID:          info.ID,
			ServiceName: info.ServiceName,
			ServiceDesc: info.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         counter.QPS,
			Qpd:         counter.TotalCount,
			TotalNode:   len(ipList),
		}
		list = append(list, item)
	}

	public.ResponseSuccess(c, &dto.ServiceListOutput{List: list, Total: total})
}

// ServiceDelete godoc
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /services/:id
// @Accept  json
// @Produce  json
// @Param id path string true "服务ID"
// @Success 200 {object} public.Response{} "success"
// @Router /services/{id} [DELETE]
func ServiceDelete(c *gin.Context) {
	p := &dto.ServiceDeleteInput{}
	if err := p.BindValidParam(c); err != nil {
		public.ResponseError(c, 2005, err)
		return
	}

	tx, err := lib.GetGormPool(DB_SCOPE)
	if err != nil {
		public.ResponseError(c, public.GetGormPoolErrorCode, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{ID: p.ID}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		public.ResponseError(c, 2006, err)
		return
	}

	serviceInfo.IsDelete = 1
	if err := serviceInfo.Save(c, tx); err != nil {
		public.ResponseError(c, 2007, err)
		return
	}

	public.ResponseSuccessWithoutData(c)
}

// ServiceDetail godoc
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /services/detail
// @Accept  json
// @Produce  json
// @Param id path string true "服务ID"
// @Success 200 {object} public.Response{data=dao.ServiceDetail} "success"
// @Router /services/{id} [GET]
func ServiceDetail(c *gin.Context) {

}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /services/:id/stat
// @Accept  json
// @Produce  json
// @Param id path string true "服务ID"
// @Success 200 {object} public.Response{data=dto.ServiceStatOutput} "success"
// @Router /services/{id}/stat [GET]
func ServiceStat(c *gin.Context) {

}

// ServiceAddHTTP godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @ID /services/add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} public.Response{data=string} "success"
// @Router /services/http [POST]
func ServiceAddHTTP(c *gin.Context) {
	p := &dto.ServiceAddHTTPInput{}
	if err := p.BindValidParam(c); err != nil {
		public.ResponseError(c, 2008, err)
		return
	}

	tx, err := lib.GetGormPool(DB_SCOPE)
	if err != nil {
		public.ResponseError(c, public.GetGormPoolErrorCode, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{ServiceName: p.ServiceName}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err == nil {
		public.ResponseError(c, 2009, errors.New("服务名称已存在"))
		return
	}

	public.ResponseSuccessWithoutData(c)
}

// ServiceUpdateHTTP godoc
// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags 服务管理
// @ID /services/update_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} public.Response{data=string} "success"
// @Router /services/http [PUT]
func ServiceUpdateHTTP(c *gin.Context) {

}

// ServiceAddTCP godoc
// @Summary tcp服务添加
// @Description tcp服务添加
// @Tags 服务管理
// @ID /services/add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} public.Response{data=string} "success"
// @Router /services/tcp [POST]
func ServiceAddTCP(c *gin.Context) {

}

// ServiceUpdateTCP godoc
// @Summary tcp服务更新
// @Description tcp服务更新
// @Tags 服务管理
// @ID /services/update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} public.Response{data=string} "success"
// @Router /services/tcp [PUT]
func ServiceUpdateTCP(c *gin.Context) {

}

// ServiceAddGRPC godoc
// @Summary grpc服务添加
// @Description grpc服务添加
// @Tags 服务管理
// @ID /services/add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} public.Response{data=string} "success"
// @Router /services/grpc [POST]
func ServiceAddGRPC(c *gin.Context) {

}

// ServiceUpdateGRPC godoc
// @Summary grpc服务更新
// @Description grpc服务更新
// @Tags 服务管理
// @ID /services/update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} public.Response{data=string} "success"
// @Router /services/grpc [PUT]
func ServiceUpdateGRPC(c *gin.Context) {

}
