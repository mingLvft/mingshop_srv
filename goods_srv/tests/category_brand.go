package main

import (
	"context"
	"fmt"
	"mingshop_srvs/goods_srv/proto"
)

func TestCategoryBrandList() {
	rsp, err := brandClient.CategoryBrandList(context.Background(), &proto.CategoryBrandFilterRequest{})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Data)
	fmt.Println(rsp.Total)
}

func TestGetCategoryBrandList() {
	rsp, err := brandClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: 135200,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Data)
	fmt.Println(rsp.Total)
}
