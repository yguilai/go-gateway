package dao

import (
	gin "github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yguilai/go-gateway/common/lib"
	dto "github.com/yguilai/go-gateway/dto"
	"github.com/yguilai/go-gateway/public"
	"net/http/httptest"
	"strings"
	"sync"
)

type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http规则"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp规则"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc规则"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"负载均衡信息"`
	AccessControl *AccessControl `json:"access_control" description:"请求控制信息"`
}

type ServiceManager struct {
	ServiceMap   map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	Locker       sync.RWMutex
	init         sync.Once
	err          error
}

var ServiceManagerHandler *ServiceManager

func init() {
	ServiceManagerHandler = NewServiceManager()
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}

func (s *ServiceManager) GetTcpServiceList() []*ServiceDetail {
	list := make([]*ServiceDetail, 0)
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == public.LoadTypeTCP {
			list = append(list, tempItem)
		}
	}
	return list
}

func (s *ServiceManager) GetGrpcServiceList() []*ServiceDetail {
	list := make([]*ServiceDetail, 0)
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == public.LoadTypeGRPC {
			list = append(list, tempItem)
		}
	}
	return list
}

func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*ServiceDetail, error) {
	//1、前缀匹配 /abc ==> serviceSlice.rule
	//2、域名匹配 www.test.com ==> serviceSlice.rule
	//host c.Request.Host
	//path c.Request.URL.Path
	host := c.Request.Host
	host = host[:strings.Index(host, ":")]
	path := c.Request.URL.Path
	for _, item := range s.ServiceSlice {
		if item.Info.LoadType != public.LoadTypeHTTP {
			continue
		}

		if item.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			if item.HTTPRule.Rule == host {
				return item, nil
			}
		}
		if item.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL {
			if strings.HasPrefix(path, item.HTTPRule.Rule) {
				return item, nil
			}
		}
	}
	return nil, errors.New("not matched service")
}

const MaxServiceCount = 99999

// LoadOnce 服务启动 载入数据
func (s *ServiceManager) LoadOnce() error {
	s.init.Do(func() {
		serviceInfo := &ServiceInfo{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &dto.ServiceListInput{PageNo: 1, PageSize: MaxServiceCount}
		list, _, err := serviceInfo.PageList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, item := range list {
			tmpItem := item
			serviceDetail, err := tmpItem.ServiceDetail(c, tx, &tmpItem)
			if err != nil {
				s.err = err
				return
			}
			s.ServiceMap[item.ServiceName] = serviceDetail
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}
	})
	return s.err
}
