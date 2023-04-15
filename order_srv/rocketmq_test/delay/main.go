package main

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

func main() {
	p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.1.4:9876"}))
	if err != nil {
		panic("生成producer失败")
	}

	if err = p.Start(); err != nil {
		panic("启动producer失败")
	}

	msg := primitive.NewMessage("ming", []byte("this is delay message"))
	msg.WithDelayTimeLevel(2)
	res, err := p.SendSync(context.Background(), msg)
	if err != nil {
		fmt.Printf("发送消息失败: %s\n", err)
	} else {
		fmt.Printf("发送消息成功: %s\n", res.String())
	}

	if err = p.Shutdown(); err != nil {
		panic("关闭producer失败")
	}
}
