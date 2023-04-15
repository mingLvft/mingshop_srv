package global

import (
	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
	"mingshop_srvs/order_srv/config"
	"mingshop_srvs/order_srv/proto"
)

var (
	DB                 *gorm.DB
	ServerConfig       *config.ServerConfig = &config.ServerConfig{}
	NacosConfig        *config.NacosConfig  = &config.NacosConfig{}
	RedisLock          *redsync.Redsync
	GoodsSrvClient     proto.GoodsClient
	InventorySrvClient proto.InventoryClient
)

//func init() {
//	dsn := "root:123456@tcp(127.0.0.1:3306)/shop_order_srv?charset=utf8mb4&parseTime=True&loc=Local"
//
//	//设置一个全局的logger，这个logger会在每次执行sql的时候打印出来
//	newLogger := logger.New(
//		log.New(os.Stdout, "\n", log.LstdFlags), // io writer
//		logger.Config{
//			SlowThreshold: time.Second, // 慢 SQL 阈值
//			LogLevel:      logger.Info, // Log level
//			Colorful:      true,        // 禁用彩色打印
//		},
//	)
//
//	//全局模式
//	var err error
//	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
//		NamingStrategy: schema.NamingStrategy{
//			SingularTable: true,
//		},
//		Logger: newLogger,
//	})
//	if err != nil {
//		panic(err)
//	}
//}
