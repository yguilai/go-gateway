package dao

import (
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/yguilai/go-gateway/dto"
	"github.com/yguilai/go-gateway/public"
	"time"
)

type ServiceInfo struct {
	ID          int64     `json:"id" gorm:"primary_key"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	UpdatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"更新时间"`
	CreatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"添加时间"`
	IsDelete    int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (t *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (t *ServiceInfo) ServiceDetail(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceDetail, error) {
	var err error
	if search.ServiceName == "" {
		info, err := t.Find(c, tx, search)
		if err != nil {
			return nil, err
		}
		search = info
	}
	httpRule := &HttpRule{ServiceID: search.ID}
	httpRule, err = httpRule.Find(c, tx, httpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	tcpRule := &TcpRule{ServiceID: search.ID}
	tcpRule, err = tcpRule.Find(c, tx, tcpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	grpcRule := &GrpcRule{ServiceID: search.ID}
	grpcRule, err = grpcRule.Find(c, tx, grpcRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	loadBalanceRule := &LoadBalance{ServiceID: search.ID}
	loadBalanceRule, err = loadBalanceRule.Find(c, tx, loadBalanceRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	accessControlRule := &AccessControl{ServiceID: search.ID}
	accessControlRule, err = accessControlRule.Find(c, tx, accessControlRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &ServiceDetail{
		Info:          search,
		HTTPRule:      httpRule,
		TCPRule:       tcpRule,
		GRPCRule:      grpcRule,
		LoadBalance:   loadBalanceRule,
		AccessControl: accessControlRule,
	}, nil
}

func (t *ServiceInfo) PageList(c *gin.Context, tx *gorm.DB, p *dto.ServiceListInput) ([]ServiceInfo, int64, error) {
	var (
		list   []ServiceInfo
		total  = int64(0)
		offset = (p.PageNo - 1) * p.PageSize
	)

	query := tx.SetCtx(public.GetGinTraceContext(c))
	query = query.Table(t.TableName()).Where("is_delete = 0")

	if p.Info != "" {
		q := "%" + p.Info + "%"
		query = query.Where("(service_name like ? or service_desc like ?)", q, q)
	}

	if err := query.Limit(p.PageSize).Offset(offset).Order("id desc").Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	query.Limit(p.PageSize).Offset(offset).Count(&total)
	return list, total, nil

}

func (t *ServiceInfo) Find(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	o := &ServiceInfo{}
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(o).Error
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (t *ServiceInfo) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error
}
