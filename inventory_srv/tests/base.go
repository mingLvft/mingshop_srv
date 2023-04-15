package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"mingshop_srvs/inventory_srv/proto"
	"sync"
)

var invClient proto.InventoryClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	invClient = proto.NewInventoryClient(conn)
}

func main() {
	Init()
	//for i := 421; i <= 840; i++ {
	//	TestSetInv(int32(i), 100)
	//}
	//并发情况下，库存无法正确扣减
	var wg sync.WaitGroup
	wg.Add(50)
	for i := 0; i < 50; i++ {
		//time.Sleep(200 * time.Millisecond)
		go TestSell(&wg)
	}
	wg.Wait()
	defer conn.Close()
	//TestSetInv(422, 40)
	//TestInvDetail(421)
	//TestSell()
	//TestReback()
}
