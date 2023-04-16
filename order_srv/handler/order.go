package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"math/rand"
	"mingshop_srvs/order_srv/global"
	"mingshop_srvs/order_srv/model"
	"mingshop_srvs/order_srv/proto"
	"time"
)

type OrderServer struct {
	proto.UnimplementedOrderServer
}

func GenerateOrderSn(userId int32) string {
	//订单号生成规则：年月日时分秒时间戳+用户id+两位随机数
	now := time.Now()
	rand.Seed(time.Now().UnixNano())
	orderSn := fmt.Sprintf("%d%d%d%d%d%d%d%d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Nanosecond(),
		userId, rand.Intn(90)+10,
	)
	return orderSn
}

func (o *OrderServer) CartItemList(ctx context.Context, req *proto.UserInfo) (*proto.CartItemListResponse, error) {
	//获取用户的购物车列表
	var shopCarts []model.ShoppingCart
	var rsp proto.CartItemListResponse

	if result := global.DB.Where(&model.ShoppingCart{User: req.Id}).Find(&shopCarts); result.Error != nil {
		return nil, result.Error
	} else {
		rsp.Total = int32(result.RowsAffected)
	}

	for _, shopCart := range shopCarts {
		rsp.Data = append(rsp.Data, &proto.ShopCartInfoResponse{
			Id:      shopCart.ID,
			UserId:  shopCart.User,
			GoodsId: shopCart.Goods,
			Nums:    shopCart.Nums,
			Checked: shopCart.Checked,
		})
	}
	return &rsp, nil
}

func (o *OrderServer) CreateCartItem(ctx context.Context, req *proto.CartItemRequest) (*proto.ShopCartInfoResponse, error) {
	//将商品添加到购物车 1.购物车原本没有这件商品，直接添加 2.购物车原本有这件商品，数量累加
	var shopCart model.ShoppingCart

	if result := global.DB.Where(&model.ShoppingCart{User: req.UserId, Goods: req.GoodsId}).First(&shopCart); result.RowsAffected == 1 {
		shopCart.Nums += req.Nums
	} else {
		shopCart.User = req.UserId
		shopCart.Goods = req.GoodsId
		shopCart.Nums = req.Nums
		shopCart.Checked = false
	}
	global.DB.Save(&shopCart)
	return &proto.ShopCartInfoResponse{Id: shopCart.ID}, nil
}

func (c *OrderServer) UpdateCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	//更新购物车记录，更新数量和选中状态
	var shopCart model.ShoppingCart

	if result := global.DB.Where("goods=? and user=?", req.GoodsId, req.UserId).First(&shopCart); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}

	shopCart.Checked = req.Checked
	if req.Nums > 0 {
		shopCart.Nums = req.Nums
	}
	global.DB.Save(&shopCart)
	return &emptypb.Empty{}, nil
}

func (o *OrderServer) DeleteCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	if result := global.DB.Where("goods=? and user=?", req.GoodsId, req.UserId).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}
	return &emptypb.Empty{}, nil
}

func (o *OrderServer) OrderList(ctx context.Context, req *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	var orders []model.OrderInfo
	var rsp proto.OrderListResponse

	//是后台管理系统查询 还是电商系统查询
	//如果没有req.UserId，说明是后台管理系统查询，gorm的查询条件是空的，查询所有订单
	var total int64
	result := global.DB.Model(&model.OrderInfo{}).Where(&model.OrderInfo{User: req.UserId}).Count(&total)
	rsp.Total = int32(total)

	//分页
	global.DB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Where(&model.OrderInfo{User: req.UserId}).Find(&orders)
	for _, order := range orders {
		rsp.Data = append(rsp.Data, &proto.OrderInfoResponse{
			Id:      order.ID,
			UserId:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Total:   order.OrderMount,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SingerMobile,
			AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &rsp, result.Error
}

func (o *OrderServer) OrderDetail(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	var order model.OrderInfo
	var rsp proto.OrderInfoDetailResponse

	//这个订单的id是否是当前用户的订单， 如果在web层用户传递过来一个id的订单， web层应该先查询一下订单id是否是当前用户的
	//在个人中心可以这样做，但是如果是后台管理系统，web层如果是后台管理系统 那么只传递order的id，如果是电商系统还需要一个用户的id
	//如果没有req.UserId，说明是后台管理系统查询，gorm的查询条件是空的，查询所有订单
	if result := global.DB.Where(&model.OrderInfo{BaseModel: model.BaseModel{ID: req.Id}, User: req.UserId}).First(&order); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}

	orderInfo := proto.OrderInfoResponse{}
	orderInfo.Id = order.ID
	orderInfo.UserId = order.User
	orderInfo.OrderSn = order.OrderSn
	orderInfo.PayType = order.PayType
	orderInfo.Status = order.Status
	orderInfo.Post = order.Post
	orderInfo.Total = order.OrderMount
	orderInfo.Address = order.Address
	orderInfo.Name = order.SignerName
	orderInfo.Mobile = order.SingerMobile

	rsp.OrderInfo = &orderInfo

	var orderGoods []model.OrderGoods
	if result := global.DB.Where(&model.OrderGoods{Order: order.ID}).Find(&orderGoods); result.Error != nil {
		return nil, result.Error
	}
	for _, orderGood := range orderGoods {
		rsp.Goods = append(rsp.Goods, &proto.OrderItemResponse{
			GoodsId:    orderGood.Goods,
			GoodsName:  orderGood.GoodsName,
			GoodsImage: orderGood.GoodsImage,
			GoodsPrice: orderGood.GoodsPrice,
			Nums:       orderGood.Nums,
		})
	}
	return &rsp, nil
}

type OrderListener struct {
	Code        codes.Code
	Detail      string
	ID          int32
	OrderAmount float32
}

func (o *OrderListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	var orderInfo model.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)

	var goodsIds []int32
	var shopCarts []model.ShoppingCart
	goodsNumsMap := make(map[int32]int32)
	if result := global.DB.Where(&model.ShoppingCart{User: orderInfo.User, Checked: true}).Find(&shopCarts); result.RowsAffected == 0 {
		o.Code = codes.InvalidArgument
		o.Detail = "购物车中没有选中的商品"
		return primitive.RollbackMessageState
	}

	for _, shopCart := range shopCarts {
		goodsIds = append(goodsIds, shopCart.Goods)
		goodsNumsMap[shopCart.Goods] = shopCart.Nums
	}

	//跨商品服务调用
	goods, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{Id: goodsIds})
	if err != nil {
		o.Code = codes.Internal
		o.Detail = "批量查询商品信息失败"
		return primitive.RollbackMessageState
	}

	var orderAmount float32
	var orderGoods []*model.OrderGoods
	var goodsInvInfo []*proto.GoodsInvInfo
	for _, good := range goods.Data {
		orderAmount += good.ShopPrice * float32(goodsNumsMap[good.Id])
		orderGoods = append(orderGoods, &model.OrderGoods{
			Goods:      good.Id,
			GoodsName:  good.Name,
			GoodsImage: good.GoodsFrontImage,
			GoodsPrice: good.ShopPrice,
			Nums:       goodsNumsMap[good.Id],
		})

		goodsInvInfo = append(goodsInvInfo, &proto.GoodsInvInfo{
			GoodsId: good.Id,
			Num:     goodsNumsMap[good.Id],
		})
	}

	//跨库存服务调用
	if _, err = global.InventorySrvClient.Sell(context.Background(), &proto.SellInfo{OrderSn: orderInfo.OrderSn, GoodsInfo: goodsInvInfo}); err != nil {
		statusCode := status.Convert(err).Code()
		if statusCode == codes.Internal || statusCode == codes.InvalidArgument || statusCode == codes.ResourceExhausted {
			o.Code = codes.ResourceExhausted
			o.Detail = "扣减库存失败"
			return primitive.RollbackMessageState
		}
	}
	//测试提交消息，回滚库存
	//return primitive.CommitMessageState
	//测试未知情况，进行回查
	//return primitive.UnknowState

	//生成订单
	tx := global.DB.Begin()
	orderInfo.OrderMount = orderAmount
	if result := tx.Save(&orderInfo); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "创建订单失败"
		return primitive.CommitMessageState
	}

	//订单商品信息
	o.OrderAmount = orderAmount
	o.ID = orderInfo.ID
	for _, orderGood := range orderGoods {
		orderGood.Order = orderInfo.ID
	}
	//批量插入订单商品
	if result := tx.CreateInBatches(orderGoods, 100); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "订单商品创建失败"
		return primitive.CommitMessageState
	}

	//删除购物车中已经购买的商品
	if result := tx.Where(&model.ShoppingCart{User: orderInfo.User, Checked: true}).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "购物车商品删除失败"
		return primitive.CommitMessageState
	}

	//发送延时消息
	p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.1.4:9876"}), producer.WithGroupName("mingshop-order"))
	if err != nil {
		zap.S().Errorf("生成producer失败: %s", err.Error())
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "生成producer失败"
		return primitive.CommitMessageState
	}

	if err = p.Start(); err != nil {
		zap.S().Errorf("启动producer失败: %s", err.Error())
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "启动producer失败"
		return primitive.CommitMessageState
	}

	msg = primitive.NewMessage("order_timeout", msg.Body)
	msg.WithDelayTimeLevel(2)
	_, err = p.SendSync(context.Background(), msg)
	if err != nil {
		zap.S().Errorf("发送延时消息失败: %v\n", err)
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "发送延时消息失败"
		return primitive.CommitMessageState
	}

	tx.Commit()
	o.Code = codes.OK
	return primitive.RollbackMessageState
}

func (o *OrderListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	var orderInfo model.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)

	//如果订单已经存在，说明整个流程没有问题
	if result := global.DB.Where(model.OrderInfo{OrderSn: orderInfo.OrderSn}).First(&orderInfo); result.RowsAffected == 0 {
		//这里并不能确定库存已经扣减了，所以消费者需要做好幂等性
		return primitive.CommitMessageState
	}
	return primitive.RollbackMessageState
}

func (o *OrderServer) CreateOrder(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	/**
	新建订单
	1。从购物车中获取选中的商品
	2。商品的金额自己查询 - 访问商品服务（跨微服务）
	3。库存的扣减 - 访问库存服务（跨微服务）
	4。订单的基本信息表的创建 - 订单的商品信息表的创建
	5。从购物车中删除已购买的记录
	*/
	orderListener := OrderListener{}
	p, err := rocketmq.NewTransactionProducer(
		&orderListener,
		producer.WithNameServer([]string{"192.168.1.4:9876"}),
	)
	if err != nil {
		zap.S().Errorf("生成producer失败: %s", err.Error())
		return nil, err
	}

	if err = p.Start(); err != nil {
		zap.S().Errorf("启动producer失败: %s", err.Error())
		return nil, err
	}

	order := model.OrderInfo{
		OrderSn:      GenerateOrderSn(req.UserId),
		Address:      req.Address,
		SignerName:   req.Name,
		SingerMobile: req.Mobile,
		Post:         req.Post,
		User:         req.UserId,
	}

	//应该在消息中具体指明每一个订单的具体的商品的扣减情况
	jsonString, _ := json.Marshal(order)

	_, err = p.SendMessageInTransaction(context.Background(), primitive.NewMessage("order_reback", jsonString))
	if err != nil {
		fmt.Printf("发送消息失败: %v\n", err)
		return nil, status.Error(codes.Internal, "发送消息失败")
	}
	if orderListener.Code != codes.OK {
		return nil, status.Error(orderListener.Code, orderListener.Detail)
	}

	return &proto.OrderInfoResponse{Id: orderListener.ID, OrderSn: order.OrderSn, Total: orderListener.OrderAmount}, nil
}

//func (o *OrderServer) CreateOrder(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
//	/**
//	新建订单
//	1。从购物车中获取选中的商品
//	2。商品的金额自己查询 - 访问商品服务（跨微服务）
//	3。库存的扣减 - 访问库存服务（跨微服务）
//	4。订单的基本信息表的创建 - 订单的商品信息表的创建
//	5。从购物车中删除已购买的记录
//	*/
//	p, err := rocketmq.NewTransactionProducer(
//		&OrderListener{},
//		producer.WithNameServer([]string{"192.168.1.4:9876"}),
//	)
//	if err != nil {
//		zap.S().Errorf("生成producer失败: %s", err.Error())
//		return nil, err
//	}
//
//	if err = p.Start(); err != nil {
//		zap.S().Errorf("启动producer失败: %s", err.Error())
//		return nil, err
//	}
//
//	res, err := p.SendMessageInTransaction(context.Background(), primitive.NewMessage("transTopic_ming", []byte("this is transaction message")))
//	if err != nil {
//		fmt.Printf("发送消息失败: %v\n", err)
//	} else {
//		fmt.Printf("发送消息成功: %s\n", res.String())
//	}
//
//	var goodsIds []int32
//	var shopCarts []model.ShoppingCart
//	goodsNumsMap := make(map[int32]int32)
//	if result := global.DB.Where(&model.ShoppingCart{User: req.UserId, Checked: true}).Find(&shopCarts); result.RowsAffected == 0 {
//		return nil, status.Errorf(codes.NotFound, "购物车中没有选中的商品")
//	}
//
//	for _, shopCart := range shopCarts {
//		goodsIds = append(goodsIds, shopCart.Goods)
//		goodsNumsMap[shopCart.Goods] = shopCart.Nums
//	}
//
//	//跨商品服务调用
//	goods, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{Id: goodsIds})
//	if err != nil {
//		return nil, status.Errorf(codes.Internal, "批量查询商品信息失败")
//	}
//
//	var orderAmount float32
//	var orderGoods []*model.OrderGoods
//	var goodsInvInfo []*proto.GoodsInvInfo
//	for _, good := range goods.Data {
//		orderAmount += good.ShopPrice * float32(goodsNumsMap[good.Id])
//		orderGoods = append(orderGoods, &model.OrderGoods{
//			Goods:      good.Id,
//			GoodsName:  good.Name,
//			GoodsImage: good.GoodsFrontImage,
//			GoodsPrice: good.ShopPrice,
//			Nums:       goodsNumsMap[good.Id],
//		})
//
//		goodsInvInfo = append(goodsInvInfo, &proto.GoodsInvInfo{
//			GoodsId: good.Id,
//			Num:     goodsNumsMap[good.Id],
//		})
//	}
//
//	//跨库存服务调用
//	if _, err = global.InventorySrvClient.Sell(context.Background(), &proto.SellInfo{GoodsInfo: goodsInvInfo}); err != nil {
//		return nil, status.Errorf(codes.ResourceExhausted, "扣减库存失败")
//	}
//
//	//生成订单
//	tx := global.DB.Begin()
//	order := model.OrderInfo{
//		OrderSn:      GenerateOrderSn(req.UserId),
//		OrderMount:   orderAmount,
//		Address:      req.Address,
//		SignerName:   req.Name,
//		SingerMobile: req.Mobile,
//		Post:         req.Post,
//		User:         req.UserId,
//	}
//	if result := tx.Save(&order); result.RowsAffected == 0 {
//		tx.Rollback()
//		return nil, status.Errorf(codes.Internal, "订单创建失败")
//	}
//
//	//订单商品信息
//	for _, orderGood := range orderGoods {
//		orderGood.Order = order.ID
//	}
//	if result := tx.CreateInBatches(orderGoods, 100); result.RowsAffected == 0 {
//		tx.Rollback()
//		return nil, status.Errorf(codes.Internal, "订单商品创建失败")
//	}
//
//	//删除购物车中已经购买的商品
//	if result := tx.Where(&model.ShoppingCart{User: req.UserId, Checked: true}).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
//		tx.Rollback()
//		return nil, status.Errorf(codes.Internal, "购物车商品删除失败")
//	}
//
//	tx.Commit()
//	return &proto.OrderInfoResponse{Id: order.ID, OrderSn: order.OrderSn, Total: orderAmount}, nil
//}

func (o *OrderServer) UpdateOrderStatus(ctx context.Context, req *proto.OrderStatus) (*emptypb.Empty, error) {
	if result := global.DB.Model(&model.OrderInfo{}).Where("order_sn = ?", req.OrderSn).Update("status", req.Status); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	return &emptypb.Empty{}, nil
}

func OrderTimeout(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range msgs {
		var orderInfo model.OrderInfo
		_ = json.Unmarshal(msgs[i].Body, &orderInfo)

		fmt.Printf("获取到订单超时消息: %v\n", time.Now())
		//查询订单的支付状态，如果已经支付则不做处理，如果未支付则取消订单
		var order model.OrderInfo
		if result := global.DB.Model(model.OrderInfo{}).Where(model.OrderInfo{OrderSn: orderInfo.OrderSn}).First(&order); result.RowsAffected == 0 {
			return consumer.ConsumeSuccess, nil
		}
		if order.Status != "TRADE_SUCCESS" {
			tx := global.DB.Begin()
			//归还库存，我们可以模仿order中发送一个消息到order_reback中去
			//修改订单的状态为交易已结束
			order.Status = "TRADE_CLOSED"
			tx.Save(&order)
			p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.1.4:9876"}), producer.WithGroupName("mingshop-inventory"))
			if err != nil {
				zap.S().Errorf("生成producer失败: %s", err.Error())
				tx.Rollback()
			}

			if err = p.Start(); err != nil {
				zap.S().Errorf("启动producer失败: %s", err.Error())
				tx.Rollback()
			}

			_, err = p.SendSync(context.Background(), primitive.NewMessage("order_reback", msgs[i].Body))
			if err != nil {
				tx.Rollback()
				fmt.Printf("发送消息失败: %s\n", err)
				return consumer.ConsumeRetryLater, nil
			}

			tx.Commit()
		}
	}
	return consumer.ConsumeSuccess, nil
}
