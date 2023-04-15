package handler

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"mingshop_srvs/inventory_srv/global"
	"mingshop_srvs/inventory_srv/model"
	"mingshop_srvs/inventory_srv/proto"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

func (*InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	//设置库存
	var inv model.Inventory
	global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv)
	inv.Goods = req.GoodsId
	inv.Stocks = req.Num

	global.DB.Save(&inv)
	return &emptypb.Empty{}, nil
}

func (*InventoryServer) InvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	//获取库存信息
	var inv model.Inventory
	if result := global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "没有库存信息")
	}
	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

//var m sync.Mutex

func (*InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//库存扣减
	//并发场景下，需要加锁
	tx := global.DB.Begin()
	//m.Lock() //加锁
	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory
		//悲观锁
		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
		//	tx.Rollback() //回滚
		//	return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		//}
		//乐观锁(乐观锁，失败会自动重试,不担心旧数据问题)
		//for {
		//redis分布式锁
		mutex := global.RedisLock.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁失败")
		}

		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减
		inv.Stocks -= goodInfo.Num
		tx.Save(&inv)
		tx.Commit() //提交（如果不提交就释放锁，会导致数据没有被更改，然后另一个进程就获取了锁并且继续用旧数据(获取锁和释放锁中间的逻辑执行太快，都还没到commit就释放了，被其他进程获取到了锁，然后查询了旧数据)）

		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁失败")
		}

		// update inventory set stocks=stocks-1,version=version+1 where goods = goods and version=version;
		//if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version = ?", goodInfo.GoodsId, inv.Version).Updates(model.Inventory{Stocks: inv.Stocks, Version: inv.Version + 1}); result.RowsAffected == 0 {
		//	zap.S().Info("库存扣减失败")
		//} else {
		//	break
		//}
		//}
		//tx.Save(&inv)
	}
	//tx.Commit() //提交 （悲观锁 commit的时候才会释放锁）
	//m.Unlock()  //解锁 (如果不提交就释放锁，会导致数据没有被更改，然后另一个进程就获取了锁并且继续用旧数据(获取锁和释放锁中间的逻辑执行太快，都还没到commit就释放了，被其他进程获取到了锁，然后查询了旧数据))
	return &emptypb.Empty{}, nil
}

func (*InventoryServer) Reback(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//库存归还 1. 订单超时归还 2. 订单创建失败归还  3.手动归还
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		inv.Stocks += goodInfo.Num
		tx.Save(&inv)
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}
