package tcp_proxy_router

import (
	"context"
	"fmt"
	"github.com/yguilai/go-gateway/dao"
	"github.com/yguilai/go-gateway/reverse_proxy"
	"github.com/yguilai/go-gateway/tcp_proxy_middleware"
	"github.com/yguilai/go-gateway/tcp_server"
	"log"
	"net"
)

var tcpServerList = make([]*tcp_server.TcpServer, 0)

type tcpHandler struct {
}

func (t *tcpHandler) ServeTCP(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}

func TcpServerRun() {
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()
	for _, item := range serviceList {
		tmp := item
		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.TCPRule.Port)
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}

			router := tcp_proxy_middleware.NewTcpSliceRouter()
			router.Group("/").Use(
				tcp_proxy_middleware.TCPFlowCountMiddleware(),
				tcp_proxy_middleware.TCPFlowLimitMiddleware(),
				tcp_proxy_middleware.TCPWhiteListMiddleware(),
				tcp_proxy_middleware.TCPBlackListMiddleware(),
			)

			routerHandler := tcp_proxy_middleware.NewTcpSliceRouterHandler(func(c *tcp_proxy_middleware.TcpSliceRouterContext) tcp_server.TCPHandler {
				return reverse_proxy.NewTcpLoadBalanceReverseProxy(c, rb)
			}, router)

			baseCtx := context.WithValue(context.Background(), "service", serviceDetail)
			tcpServer := &tcp_server.TcpServer{
				Addr:    addr,
				Handler: routerHandler,
				BaseCtx: baseCtx,
			}
			tcpServerList = append(tcpServerList, tcpServer)
			log.Printf(" [INFO] tcp_proxy_run %v\n", addr)
			if err := tcpServer.ListenAndServe(); err != nil && err != tcp_server.ErrServerClosed {
				log.Fatalf(" [INFO] tcp_proxy_run %v err:%v\n", addr, err)
			}
		}(tmp)
	}
}

func TcpServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf(" [INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
	}
}