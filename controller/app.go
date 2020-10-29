package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yguilai/go-gateway/common/lib"
	"github.com/yguilai/go-gateway/dao"
	"github.com/yguilai/go-gateway/dto"
	"github.com/yguilai/go-gateway/public"
	"time"
)

func RegisterAppController(r *gin.RouterGroup) {
	r.GET("", APPList)
	r.GET("/:id", APPDetail)
	r.DELETE("/:id", APPDelete)
	r.POST("", AppAdd)
	r.PUT("", AppUpdate)
	r.GET("/:id/stat", AppStatistics)
}

// APPList godoc
// @Summary 租户列表
// @Description 租户列表
// @Tags 租户管理
// @ID /apps
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query string true "每页多少条"
// @Param page_no query string true "页码"
// @Success 200 {object} public.Response{data=dto.APPListOutput} "success"
// @Router /apps [GET]
func APPList(c *gin.Context) {
	params := &dto.APPListInput{}
	if err := params.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}
	info := &dao.App{}
	list, total, err := info.APPList(c, lib.GORMDefaultPool, params)
	if err != nil {
		public.ResponseError(c, 2002, err)
		return
	}

	outputList := []dto.APPListItemOutput{}
	for _, item := range list {
		appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + item.AppID)
		if err != nil {
			public.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		outputList = append(outputList, dto.APPListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			Qps:      item.Qps,
			RealQpd:  appCounter.TotalCount,
			RealQps:  appCounter.QPS,
		})
	}
	output := dto.APPListOutput{
		List:  outputList,
		Total: total,
	}
	public.ResponseSuccess(c, output)
	return
}

// APPDetail godoc
// @Summary 租户详情
// @Description 租户详情
// @Tags 租户管理
// @ID /apps/detail
// @Accept  json
// @Produce  json
// @Param id path string true "租户ID"
// @Success 200 {object} public.Response{data=dao.App} "success"
// @Router /apps/{id} [get]
func APPDetail(c *gin.Context) {
	params := &dto.APPDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
	}
	detail, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		public.ResponseError(c, 2002, err)
		return
	}
	public.ResponseSuccess(c, detail)
	return
}

// APPDelete godoc
// @Summary 租户删除
// @Description 租户删除
// @Tags 租户管理
// @ID /apps/delete
// @Accept  json
// @Produce  json
// @Param id path string true "租户ID"
// @Success 200 {object} public.Response{data=string} "success"
// @Router /apps/{id} [delete]
func APPDelete(c *gin.Context) {
	params := &dto.APPDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
	}
	info, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		public.ResponseError(c, 2002, err)
		return
	}
	info.IsDelete = 1
	if err := info.Save(c, lib.GORMDefaultPool); err != nil {
		public.ResponseError(c, 2003, err)
		return
	}
	public.ResponseSuccess(c, "")
	return
}

// AppAdd godoc
// @Summary 租户添加
// @Description 租户添加
// @Tags 租户管理
// @ID /apps/add
// @Accept  json
// @Produce  json
// @Param body body dto.APPAddHttpInput true "body"
// @Success 200 {object} public.Response{data=string} "success"
// @Router /apps [post]
func AppAdd(c *gin.Context) {
	params := &dto.APPAddHttpInput{}
	if err := params.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	//验证app_id是否被占用
	search := &dao.App{
		AppID: params.AppID,
	}
	if _, err := search.Find(c, lib.GORMDefaultPool, search); err == nil {
		public.ResponseError(c, 2002, errors.New("租户ID被占用，请重新输入"))
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	tx := lib.GORMDefaultPool
	info := &dao.App{
		AppID:    params.AppID,
		Name:     params.Name,
		Secret:   params.Secret,
		WhiteIPS: params.WhiteIPS,
		Qps:      params.Qps,
		Qpd:      params.Qpd,
	}
	if err := info.Save(c, tx); err != nil {
		public.ResponseError(c, 2003, err)
		return
	}
	public.ResponseSuccess(c, "")
	return
}

// AppUpdate godoc
// @Summary 租户更新
// @Description 租户更新
// @Tags 租户管理
// @ID /apps/update
// @Accept  json
// @Produce  json
// @Param body body dto.APPUpdateHttpInput true "body"
// @Success 200 {object} public.Response{data=string} "success"
// @Router /apps [put]
func AppUpdate(c *gin.Context) {
	params := &dto.APPUpdateHttpInput{}
	if err := params.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
	}
	info, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		public.ResponseError(c, 2002, err)
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	info.Name = params.Name
	info.Secret = params.Secret
	info.WhiteIPS = params.WhiteIPS
	info.Qps = params.Qps
	info.Qpd = params.Qpd
	if err := info.Save(c, lib.GORMDefaultPool); err != nil {
		public.ResponseError(c, 2003, err)
		return
	}
	public.ResponseSuccess(c, "")
	return
}

// AppStatistics godoc
// @Summary 租户统计
// @Description 租户统计
// @Tags 租户管理
// @ID /apps/stat
// @Accept  json
// @Produce  json
// @Param id path string true "租户ID"
// @Success 200 {object} public.Response{data=dto.StatisticsOutput} "success"
// @Router /apps/{id}/stat [get]
func AppStatistics(c *gin.Context) {
	params := &dto.APPDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		public.ResponseError(c, 2001, err)
		return
	}

	search := &dao.App{
		ID: params.ID,
	}
	detail, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		public.ResponseError(c, 2002, err)
		return
	}

	//今日流量全天小时级访问统计
	todayStat := []int64{}
	counter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + detail.AppID)
	if err != nil {
		public.ResponseError(c, 2002, err)
		c.Abort()
		return
	}
	currentTime:= time.Now()
	for i := 0; i <= time.Now().In(lib.TimeLocation).Hour(); i++ {
		dateTime:=time.Date(currentTime.Year(),currentTime.Month(),currentTime.Day(),i,0,0,0,lib.TimeLocation)
		hourData,_:=counter.GetHourData(dateTime)
		todayStat = append(todayStat, hourData)
	}

	//昨日流量全天小时级访问统计
	yesterdayStat := []int64{}
	yesterTime:= currentTime.Add(-1*time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime:=time.Date(yesterTime.Year(),yesterTime.Month(),yesterTime.Day(),i,0,0,0,lib.TimeLocation)
		hourData,_:=counter.GetHourData(dateTime)
		yesterdayStat = append(yesterdayStat, hourData)
	}
	stat := dto.StatisticsOutput{
		Today:     todayStat,
		Yesterday: yesterdayStat,
	}
	public.ResponseSuccess(c, stat)
	return
}
