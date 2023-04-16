package main

import (
	"flag"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"mingshop_srvs/order_srv/global"
	"mingshop_srvs/order_srv/handler"
	"mingshop_srvs/order_srv/initalize"
	"mingshop_srvs/order_srv/proto"
	"mingshop_srvs/order_srv/utils"
	"mingshop_srvs/order_srv/utils/register/consul"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	//0.0.0.0 是对外开放，说明80端口外面可以访问；127.0.0.1 是只能本机访问，外面访问不了此端口
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	//Port := flag.Int("port", 0, "端口号")
	Port := flag.Int("port", 50051, "端口号")

	//初始化
	initalize.InitLogger()
	initalize.InitConfig()
	initalize.InitDB()
	initalize.InitRedisLock()
	initalize.InitSrvConn()
	zap.S().Infof("nacos配置信息：%v", global.ServerConfig)

	flag.Parse()
	zap.S().Info("ip:", *IP)
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Info("port:", *Port)

	server := grpc.NewServer()
	proto.RegisterOrderServer(server, &handler.OrderServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("fialed to listen: " + err.Error())
	}
	//注册健康服务检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	//服务注册
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	err = register_client.Register(global.ServerConfig.Host, *Port, global.ServerConfig.Name, global.ServerConfig.Tags, serviceId)
	if err != nil {
		zap.S().Panic("服务注册失败：", err.Error())
	}
	zap.S().Debugf("启动服务器，监听端口：%d", *Port)

	//监听库存归还topic
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"192.168.1.4:9876"}),
		consumer.WithGroupName("mingshop-order"),
	)

	if err := c.Subscribe("order_timeout", consumer.MessageSelector{}, handler.OrderTimeout); err != nil {
		fmt.Printf("读取消息失败: %s\n", err)
	}
	_ = c.Start()
	//不能让主goroutine退出

	//启动服务
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
	_ = c.Shutdown()
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Info("注销服务失败")
	} else {
		zap.S().Info("注销服务成功")
	}

}
