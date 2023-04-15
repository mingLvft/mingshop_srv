package main

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"time"
)

type OrderListener struct{}

func (o *OrderListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	fmt.Println("开始执行本地逻辑")
	time.Sleep(time.Second * 3)
	fmt.Println("本地逻辑执行完毕")
	return primitive.UnknowState
}

func (o *OrderListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	fmt.Println("rocketmq的消息回查")
	time.Sleep(time.Second * 15)
	return primitive.CommitMessageState
}

func main() {
	p, err := rocketmq.NewTransactionProducer(
		&OrderListener{},
		producer.WithNameServer([]string{"192.168.1.4:9876"}),
	)
	if err != nil {
		panic("生成producer失败")
	}

	if err = p.Start(); err != nil {
		panic("启动producer失败")
	}

	res, err := p.SendMessageInTransaction(context.Background(), primitive.NewMessage("transTopic_ming", []byte("this is transaction message")))
	if err != nil {
		fmt.Printf("发送消息失败: %v\n", err)
	} else {
		fmt.Printf("发送消息成功: %s\n", res.String())
	}

	time.Sleep(time.Hour)
	if err = p.Shutdown(); err != nil {
		panic("关闭producer失败")
	}
}
