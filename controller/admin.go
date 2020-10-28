package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/yguilai/go-gateway/common/lib"
	"github.com/yguilai/go-gateway/dao"
	"github.com/yguilai/go-gateway/dto"
	"github.com/yguilai/go-gateway/public"
	"time"
)

const DB_SCOPE = "default"

func RegsiterAdminSignController(r *gin.RouterGroup)  {
	r.POST("/in", AdminLogin)
	r.POST("/out", AdminLogout)
}

func RegisterAdminController(r *gin.RouterGroup) {
	r.GET("/info", AdminInfo)
	r.POST("/update-pwd", AdminUpdatePwd)
}

// AdminLogin godoc
// @Summary 管理员登录
// @Description 管理员登录
// @Tags 管理员接口
// @ID /sign/in
// @Accept json
// @Produce json
// @Param body body dto.AdminInput true "body"
// @Success 200 {object} public.Response{data=dto.AdminOutput} "success"
// @Router /sign/in [POST]
func AdminLogin(c *gin.Context) {
	p := &dto.AdminInput{}

	if err := p.BindValidParam(c); err != nil {
		public.ResponseError(c, 1001, err)
		return
	}

	admin := &dao.Admin{}
	tx, err := lib.GetGormPool(DB_SCOPE)
	if err != nil {
		public.ResponseError(c, 1002, err)
		return
	}
	admin, err = admin.LoginCheck(c, tx, p)
	if err != nil {
		public.ResponseError(c, 1003, err)
		return
	}

	si := &dto.AdminSessionInfo{
		Id:        admin.Id,
		Username:  admin.UserName,
		LoginTime: time.Now(),
	}
	sBytes, err := json.Marshal(si)
	if err != nil {
		public.ResponseError(c, 1004, err)
		return
	}
	s := sessions.Default(c)
	s.Set(public.AdminSessionInfoKey, string(sBytes))
	s.Save()

	r := &dto.AdminOutput{Token: admin.UserName}
	public.ResponseSuccess(c, r)
}

// AdminLogout godoc
// @Summary 管理员注销登录
// @Description 管理员注销登录
// @Tags 管理员接口
// @ID /sign/out
// @Accept json
// @Produce json
// @Success 200 {object} public.Response{} "success"
// @Router /sign/out [POST]
func AdminLogout(c *gin.Context) {
	s := sessions.Default(c)
	s.Delete(public.AdminSessionInfoKey)
	s.Save()
	public.ResponseSuccess(c, nil)
}

// AdminInfo godoc
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员接口
// @ID /admin/info
// @Accept json
// @Produce json
// @Success 200 {object} public.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/info [GET]
func AdminInfo(c *gin.Context) {
	var admin dto.AdminInfoOutput
	s := sessions.Default(c)
	// SessionAuthMiddleware中间件已经校验这个值, 这里不需要再次校验
	adminInfo := s.Get(public.AdminSessionInfoKey).(string)
	err := json.Unmarshal([]byte(adminInfo), &admin)
	if err != nil {
		public.ResponseError(c, 1005, err)
		return
	}
	admin.Avatar = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
	admin.Introduction = "super administrator"
	admin.Roles = []string{"admin"}
	public.ResponseSuccess(c, admin)
}

// AdminUpdatePwd godoc
// @Summary 管理员修改密码
// @Description 管理员修改密码
// @Tags 管理员接口
// @ID /admin/update_pwd
// @Accept json
// @Produce json
// @Param body body dto.UpdatePwdInput true "body"
// @Success 200 {object} public.Response{} "success"
// @Router /admin/update_pwd [POST]
func AdminUpdatePwd(c *gin.Context)  {
	p := &dto.UpdatePwdInput{}

	if err := p.BindValidParam(c); err != nil {
		public.ResponseError(c, 1006, err)
		return
	}

	s := sessions.Default(c)
	adminInfo := s.Get(public.AdminSessionInfoKey)
	admin := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(adminInfo)), admin); err != nil {
		public.ResponseError(c, 1007, err)
		return
	}

	tx, err := lib.GetGormPool(DB_SCOPE)
	if err != nil {
		public.ResponseError(c, 1008, err)
		return
	}

	info := &dao.Admin{}
	info, err = info.Find(c, tx, &dao.Admin{Id: admin.Id})
	if err != nil {
		public.ResponseError(c, 1009, err)
		return
	}

	//生成新密码 saltPassword
	saltPassword := public.GenSaltPwd(info.Salt, p.Password)
	info.Password = saltPassword

	if err := info.Save(c, tx); err != nil {
		public.ResponseError(c, 1010, err)
		return
	}
	public.ResponseSuccess(c, nil)
}