package main

import (
	"context"
	"fmt"
	"mingshop_srvs/goods_srv/proto"
)

func TestGetUserList() {
	rsp, err := brandClient.BrandList(context.Background(), &proto.BrandFilterRequest{})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	for _, brand := range rsp.Data {
		fmt.Println(brand.Name)
	}

}
