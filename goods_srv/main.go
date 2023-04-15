package main

import (
	"flag"
	"fmt"
	"github.com/hashicorp/consul/api"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"mingshop_srvs/goods_srv/global"
	"mingshop_srvs/goods_srv/handler"
	"mingshop_srvs/goods_srv/initalize"
	"mingshop_srvs/goods_srv/proto"
	"mingshop_srvs/goods_srv/utils"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	//0.0.0.0 是对外开放，说明80端口外面可以访问；127.0.0.1 是只能本机访问，外面访问不了此端口
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	//Port := flag.Int("port", 0, "端口号")
	Port := flag.Int("port", 50052, "端口号")

	//初始化
	initalize.InitLogger()
	initalize.InitConfig()
	initalize.InitDB()
	initalize.InitEs()
	zap.S().Infof("nacos配置信息：%v", global.ServerConfig)

	flag.Parse()
	zap.S().Info("ip:", *IP)
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Info("port:", *Port)

	server := grpc.NewServer()
	proto.RegisterGoodsServer(server, &handler.GoodsServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("fialed to listen: " + err.Error())
	}
	//注册健康服务检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	//服务注册
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", global.ServerConfig.ConsulInfo.Host,
		global.ServerConfig.ConsulInfo.Port)
	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	//生成检查对象
	check := &api.AgentServiceCheck{
		//GRPC:                           fmt.Sprintf("127.0.0.1:%d", *Port),
		//GRPC:                           fmt.Sprintf("192.168.3.31:%d", *Port),
		GRPC:                           fmt.Sprintf("%s:%d", global.ServerConfig.Host, *Port),
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "15s",
	}
	//生成注册对象
	registration := new(api.AgentServiceRegistration)
	registration.Name = global.ServerConfig.Name
	serviceID := fmt.Sprintf("%s", uuid.NewV4())
	registration.ID = serviceID
	registration.Port = *Port
	registration.Tags = global.ServerConfig.Tags
	registration.Address = global.ServerConfig.Host
	registration.Check = check
	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}

	go func() {
		err = server.Serve(lis)
		if err != nil {
			panic("failed to serve: " + err.Error())
		}
	}()

	//接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = client.Agent().ServiceDeregister(serviceID); err != nil {
		zap.S().Info("注销服务失败")
	}
	zap.S().Info("注销服务成功")

}
