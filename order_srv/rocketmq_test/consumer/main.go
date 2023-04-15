package main

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"time"
)

func main() {
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"192.168.1.4:9876"}),
		consumer.WithGroupName("mingshop"),
	)

	err := c.Subscribe("ming", consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i, _ := range msgs {
			fmt.Printf("收到消息: %v\n", msgs[i])
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		fmt.Printf("读取消息失败: %s\n", err)
	}
	_ = c.Start()
	time.Sleep(time.Hour)
	_ = c.Shutdown()
}
