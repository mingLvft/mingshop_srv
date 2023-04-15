package model

//type Stock struct {
//	BaseModel
//	Name    string `gorm:"type:varchar(100);comment:'仓库名称'"`
//	Address string `gorm:"type:varchar(100);comment:'仓库地址'"`
//}

type Inventory struct {
	BaseModel
	Goods   int32 `gorm:"type:int;index;comment:'商品id'"`
	Stocks  int32 `gorm:"type:int;comment:'库存数量'"`
	Version int32 `gorm:"type:int;comment:'版本号(乐观锁)'"`
}

//type InventoryHistory struct {
//	user   int32
//	goods  int32
//	nums   int32
//	order  int32
//	status int32 // 1表示库存是预扣减，幂等性， 2表示已经支付
//}
