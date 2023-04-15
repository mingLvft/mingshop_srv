package main

import (
	"context"
	"mingshop_srvs/userop_srv/proto"
)

func TestAddressList() {
	_, err := addressClient.GetAddressList(context.Background(), &proto.AddressRequest{
		UserId: 32,
	})
	if err != nil {
		panic(err)
	}
}

func TestMessageList() {
	_, err := messageClient.MessageList(context.Background(), &proto.MessageRequest{
		UserId: 32,
	})
	if err != nil {
		panic(err)
	}
}

func TestUserFav() {
	_, err := userFavClient.GetFavList(context.Background(), &proto.UserFavRequest{
		UserId: 32,
	})
	if err != nil {
		panic(err)
	}
}
