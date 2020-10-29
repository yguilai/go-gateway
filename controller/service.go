package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yguilai/go-gateway/common/lib"
	"github.com/yguilai/go-gateway/dao"
	"github.com/yguilai/go-gateway/dto"
	"github.com/yguilai/go-gateway/public"
	"strings"
	"time"
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
	p := &dto.ServiceDetailInput{}
	if err := p.BindValidParam(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	tx, err := lib.GetGormPool(DB_SCOPE)
	if err != nil {
		public.ResponseError(c, public.GetGormPoolErrorCode, err)
		return
	}

	info := &dao.ServiceInfo{ID: p.ID}
	info, err = info.Find(c, tx, info)
	if err != nil {
		public.ResponseError(c, 2002, err)
		return
	}
	out, err := info.ServiceDetail(c, tx, info)
	if err != nil {
		public.ResponseError(c, 2003, errors.New("服务不存在"))
		return
	}
	public.ResponseSuccess(c, out)
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
	p := &dto.ServiceDeleteInput{}
	if err := p.BindValidParam(c); err != nil {
		public.ResponseError(c, 2000, err)
		return
	}

	tx, err := lib.GetGormPool(DB_SCOPE)
	if err != nil {
		public.ResponseError(c, public.GetGormPoolErrorCode, err)
	}

	info := &dao.ServiceInfo{ID: p.ID}
	detail, err := info.ServiceDetail(c, tx, info)
	if err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	counter, err := public.FlowCounterHandler.GetCounter(public.FlowServicePrefix + detail.Info.ServiceName)
	if err != nil {
		public.ResponseError(c, 2002, err)
		return
	}

	todayList := make([]int64, 0)
	currentTime := time.Now()
	for i := 0; i < currentTime.Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayList = append(todayList, hourData)
	}

	yesterdayList := make([]int64, 0)
	yesterdayTime := currentTime.Add(-1 * time.Hour * 24)
	for i := 0; i < 23; i++ {
		dateTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}

	public.ResponseSuccess(c, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
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

	tx = tx.Begin()

	serviceInfo := &dao.ServiceInfo{ServiceName: p.ServiceName}

	if _, err := serviceInfo.Find(c, tx, serviceInfo); err == nil {
		tx.Rollback()
		public.ResponseError(c, 2009, errors.New("服务名称已存在"))
		return
	}

	httpRule := &dao.HttpRule{RuleType: p.RuleType, Rule: p.Rule}
	if _, err := httpRule.Find(c, tx, httpRule); err == nil {
		tx.Rollback()
		public.ResponseError(c, 2010, errors.New("接入的前缀或域名已存在"))
		return
	}

	if len(strings.Split(p.IpList, "\n")) != len(strings.Split(p.WeightList, "\n")) {
		tx.Rollback()
		public.ResponseError(c, 2011, errors.New("IP与权重列表数量不一致"))
		return
	}

	s := &dao.ServiceInfo{
		ServiceName: p.ServiceName,
		ServiceDesc: p.ServiceDesc,
	}

	if err := s.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2012, err)
		return
	}

	httpR := &dao.HttpRule{
		ServiceID:      s.ID,
		RuleType:       p.RuleType,
		Rule:           p.Rule,
		NeedHttps:      p.NeedHttps,
		NeedWebsocket:  p.NeedWebsocket,
		NeedStripUri:   p.NeedStripUri,
		UrlRewrite:     p.UrlRewrite,
		HeaderTransfor: p.HeaderTransfor,
	}
	if err := httpR.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2013, err)
		return
	}

	ac := &dao.AccessControl{
		ServiceID:         s.ID,
		OpenAuth:          p.OpenAuth,
		BlackList:         p.BlackList,
		WhiteList:         p.WhiteList,
		ClientIPFlowLimit: p.ClientipFlowLimit,
		ServiceFlowLimit:  p.ServiceFlowLimit,
	}
	if err := ac.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2014, err)
		return
	}

	lb := &dao.LoadBalance{
		ServiceID:              s.ID,
		RoundType:              p.RoundType,
		IpList:                 p.IpList,
		WeightList:             p.WeightList,
		UpstreamConnectTimeout: p.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  p.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    p.UpstreamIdleTimeout,
		UpstreamMaxIdle:        p.UpstreamMaxIdle,
	}

	if err := lb.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2015, err)
		return
	}
	tx.Commit()
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
	p := &dto.ServiceUpdateHTTPInput{}
	if err := p.BindValidParam(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	if len(strings.Split(p.IpList, "\n")) != len(strings.Split(p.WeightList, "\n")) {
		public.ResponseError(c, 2002, errors.New("IP与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool(DB_SCOPE)
	if err != nil {
		public.ResponseError(c, public.GetGormPoolErrorCode, err)
		return
	}

	tx = tx.Begin()

	serviceInfo := &dao.ServiceInfo{ID: p.ID}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		public.ResponseError(c, 2003, errors.New("服务不存在"))
		return
	}

	detail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		public.ResponseError(c, 2004, errors.New("服务不存在"))
		return
	}

	info := detail.Info
	info.ServiceDesc = p.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2005, err)
		return
	}

	httpRule := detail.HTTPRule
	httpRule.NeedHttps = p.NeedHttps
	httpRule.NeedStripUri = p.NeedStripUri
	httpRule.NeedWebsocket = p.NeedWebsocket
	httpRule.UrlRewrite = p.UrlRewrite
	httpRule.HeaderTransfor = p.HeaderTransfor
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2006, err)
		return
	}

	ac := detail.AccessControl
	ac.OpenAuth = p.OpenAuth
	ac.BlackList = p.BlackList
	ac.WhiteList = p.WhiteList
	ac.ClientIPFlowLimit = p.ClientipFlowLimit
	ac.ServiceFlowLimit = p.ServiceFlowLimit
	if err := ac.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2007, err)
		return
	}

	lb := detail.LoadBalance
	lb.RoundType = p.RoundType
	lb.IpList = p.IpList
	lb.WeightList = p.WeightList
	lb.UpstreamConnectTimeout = p.UpstreamConnectTimeout
	lb.UpstreamHeaderTimeout = p.UpstreamHeaderTimeout
	lb.UpstreamIdleTimeout = p.UpstreamIdleTimeout
	lb.UpstreamMaxIdle = p.UpstreamMaxIdle
	if err := lb.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2008, err)
		return
	}
	tx.Commit()
	public.ResponseSuccessWithoutData(c)
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
	p := &dto.ServiceAddTcpInput{}
	if err := p.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	searchInfo := &dao.ServiceInfo{
		ServiceName: p.ServiceName,
		IsDelete:    0,
	}
	if _, err := searchInfo.Find(c, lib.GORMDefaultPool, searchInfo); err != nil {
		public.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	tcpRuleSearch := &dao.TcpRule{
		Port: p.Port,
	}
	if _, err := tcpRuleSearch.Find(c, lib.GORMDefaultPool, tcpRuleSearch); err != nil {
		public.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: p.Port,
	}
	if _, err := grpcRuleSearch.Find(c, lib.GORMDefaultPool, grpcRuleSearch); err == nil {
		public.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(p.IpList, ",")) != len(strings.Split(p.WeightList, ",")) {
		public.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: p.ServiceName,
		ServiceDesc: p.ServiceDesc,
	}

	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2006, err)
		return
	}

	lb := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  p.RoundType,
		IpList:     p.IpList,
		WeightList: p.WeightList,
		ForbidList: p.ForbidList,
	}
	if err := lb.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2007, err)
		return
	}

	ac := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          p.OpenAuth,
		BlackList:         p.BlackList,
		WhiteList:         p.WhiteList,
		WhiteHostName:     p.WhiteHostName,
		ClientIPFlowLimit: p.ClientIPFlowLimit,
		ServiceFlowLimit:  p.ServiceFlowLimit,
	}
	if err := ac.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2008, err)
		return
	}

	tcp := &dao.TcpRule{
		ServiceID: info.ID,
		Port:      p.Port,
	}
	if err := tcp.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2009, err)
		return
	}

	tx.Commit()
	public.ResponseSuccessWithoutData(c)
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
	p := &dto.ServiceUpdateTcpInput{}
	if err := p.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(p.IpList, ",")) != len(strings.Split(p.WeightList, ",")) {
		public.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	serviceInfo := &dao.ServiceInfo{
		ID: p.ID,
	}
	detail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		public.ResponseError(c, 2003, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = p.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2004, err)
		return
	}

	lb := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		lb = detail.LoadBalance
	}
	lb.ServiceID = info.ID
	lb.RoundType = p.RoundType
	lb.IpList = p.IpList
	lb.WeightList = p.WeightList
	lb.ForbidList = p.ForbidList
	if err := lb.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2005, err)
		return
	}

	ac := &dao.AccessControl{}
	if detail.AccessControl != nil {
		ac = detail.AccessControl
	}
	ac.ServiceID = info.ID
	ac.OpenAuth = p.OpenAuth
	ac.BlackList = p.BlackList
	ac.WhiteList = p.WhiteList
	ac.WhiteHostName = p.WhiteHostName
	ac.ClientIPFlowLimit = p.ClientIPFlowLimit
	ac.ServiceFlowLimit = p.ServiceFlowLimit
	if err := ac.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2006, err)
		return
	}

	tcpRule := &dao.TcpRule{}
	if detail.TCPRule != nil {
		tcpRule = detail.TCPRule
	}
	tcpRule.ServiceID = info.ID
	tcpRule.Port = p.Port
	if err := tcpRule.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2007, err)
		return
	}

	tx.Commit()
	public.ResponseSuccessWithoutData(c)
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
	p := &dto.ServiceAddGrpcInput{}
	if err := p.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: p.ServiceName,
		IsDelete:    0,
	}
	if _, err := infoSearch.Find(c, lib.GORMDefaultPool, infoSearch); err == nil {
		public.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: p.Port,
	}
	if _, err := tcpRuleSearch.Find(c, lib.GORMDefaultPool, tcpRuleSearch); err == nil {
		public.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: p.Port,
	}
	if _, err := grpcRuleSearch.Find(c, lib.GORMDefaultPool, grpcRuleSearch); err == nil {
		public.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(p.IpList, ",")) != len(strings.Split(p.WeightList, ",")) {
		public.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: p.ServiceName,
		ServiceDesc: p.ServiceDesc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2006, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  p.RoundType,
		IpList:     p.IpList,
		WeightList: p.WeightList,
		ForbidList: p.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2007, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          p.OpenAuth,
		BlackList:         p.BlackList,
		WhiteList:         p.WhiteList,
		WhiteHostName:     p.WhiteHostName,
		ClientIPFlowLimit: p.ClientIPFlowLimit,
		ServiceFlowLimit:  p.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2009, err)
		return
	}

	grpcRule := &dao.GrpcRule{
		ServiceID:      info.ID,
		Port:           p.Port,
		HeaderTransfor: p.HeaderTransfor,
	}
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2008, err)
		return
	}

	tx.Commit()
	public.ResponseSuccessWithoutData(c)
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
	p := &dto.ServiceUpdateGrpcInput{}
	if err := p.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(p.IpList, ",")) != len(strings.Split(p.WeightList, ",")) {
		public.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		ID: p.ID,
	}
	detail, err := service.ServiceDetail(c, lib.GORMDefaultPool, service)
	if err != nil {
		public.ResponseError(c, 2003, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = p.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2004, err)
		return
	}

	lb := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		lb = detail.LoadBalance
	}
	lb.ServiceID = info.ID
	lb.RoundType = p.RoundType
	lb.IpList = p.IpList
	lb.WeightList = p.WeightList
	lb.ForbidList = p.ForbidList
	if err := lb.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2005, err)
		return
	}

	grpcRule := &dao.GrpcRule{}
	if detail.GRPCRule != nil {
		grpcRule = detail.GRPCRule
	}
	grpcRule.ServiceID = info.ID
	//grpcRule.Port = p.Port
	grpcRule.HeaderTransfor = p.HeaderTransfor
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2006, err)
		return
	}

	ac := &dao.AccessControl{}
	if detail.AccessControl != nil {
		ac = detail.AccessControl
	}
	ac.ServiceID = info.ID
	ac.OpenAuth = p.OpenAuth
	ac.BlackList = p.BlackList
	ac.WhiteList = p.WhiteList
	ac.WhiteHostName = p.WhiteHostName
	ac.ClientIPFlowLimit = p.ClientIPFlowLimit
	ac.ServiceFlowLimit = p.ServiceFlowLimit
	if err := ac.Save(c, tx); err != nil {
		tx.Rollback()
		public.ResponseError(c, 2007, err)
		return
	}
	tx.Commit()
	public.ResponseSuccessWithoutData(c)
}
