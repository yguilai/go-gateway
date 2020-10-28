package dao

import (
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yguilai/go-gateway/dto"
	"github.com/yguilai/go-gateway/public"
	"time"
)

type Admin struct {
	Id        int       `json:"id" gorm:"primary_key"`
	UserName  string    `json:"user_name" gorm:"column:user_name"`
	Salt      string    `json:"salt" gorm:"column:salt"`
	Password  string    `json:"password" gorm:"column:password"`
	CreateAt  time.Time `json:"create_at" gorm:"column:create_at"`
	UpdateAt  time.Time `json:"update_at" gorm:"column:update_at"`
	IsDeleted int       `json:"is_deleted" gorm:"column:is_deleted"`
}

func (t *Admin) TableName() string {
	return "gateway_admin"
}

func (t *Admin) LoginCheck(c *gin.Context, tx *gorm.DB, p *dto.AdminInput) (*Admin, error) {
	adminInfo, err := t.Find(c, tx, &Admin{UserName: p.Username, IsDeleted: 0})
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	pwd := public.GenSaltPwd(adminInfo.Salt, p.Password)
	if adminInfo.Password != pwd {
		return nil, errors.New("密码错误")
	}
	return adminInfo, nil
}

func (t *Admin) Find(c *gin.Context, tx *gorm.DB, search *Admin) (*Admin, error) {
	o := &Admin{}
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(o).Error
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (t *Admin) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error
}
