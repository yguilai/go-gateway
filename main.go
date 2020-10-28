package main

import (
	"github.com/yguilai/go-gateway/common/lib"
	"github.com/yguilai/go-gateway/router"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"})
	if err != nil {
		log.Println(err)
	}
	defer lib.Destroy()
	router.HttpServerRun()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	router.HttpServerStop()
}
