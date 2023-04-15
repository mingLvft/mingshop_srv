package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"mingshop_srvs/goods_srv/proto"
)

func TestGetCategoryList() {
	rsp, err := brandClient.GetAllCategorysList(context.Background(), &empty.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	fmt.Println(rsp.JsonData)
}

func TestGetSubCategoryList() {
	rsp, err := brandClient.GetSubCategory(context.Background(), &proto.CategoryListRequest{
		Id: 136678,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.SubCategorys)
}
