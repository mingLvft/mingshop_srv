package main

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"mingshop_srvs/goods_srv/global"
	"mingshop_srvs/goods_srv/model"
	"os"
	"strconv"
	"time"
)

func main() {
	//dsn := "root:123456@tcp(127.0.0.1:3306)/mingshop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
	//
	////设置一个全局的logger，这个logger会在每次执行sql的时候打印出来
	//newLogger := logger.New(
	//	log.New(os.Stdout, "\n", log.LstdFlags), // io writer
	//	logger.Config{
	//		SlowThreshold: time.Second, // 慢 SQL 阈值
	//		LogLevel:      logger.Info, // Log level
	//		Colorful:      true,        // 禁用彩色打印
	//	},
	//)
	////
	////options := &password.Options{16, 100, 32, sha512.New}
	////salt, encodePwd := password.Encode("admin123", options)
	////newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodePwd)
	////fmt.Println(newPassword)
	//
	////全局模式
	//db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	//	NamingStrategy: schema.NamingStrategy{
	//		SingularTable: true,
	//	},
	//	Logger: newLogger,
	//})
	//if err != nil {
	//	panic(err)
	//}
	//
	////for i := 0; i < 10; i++ {
	////	user := model.User{
	////		NickName: fmt.Sprintf("ming%d", i),
	////		Mobile:   fmt.Sprintf("1333333333%d", i),
	////		Password: newPassword,
	////	}
	////	db.Save(&user)
	////}
	//
	////
	//////定义一个表结构，将表结构直接生成对应的表 -migrations
	//////迁移 schema
	//_ = db.AutoMigrate(&model.Category{}, &model.Brands{}, &model.GoodsCategoryBand{}, &model.Banner{}, &model.Goods{})
	Mysql2Es()
}

func Mysql2Es() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/mingshop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"

	//设置一个全局的logger，这个logger会在每次执行sql的时候打印出来
	newLogger := logger.New(
		log.New(os.Stdout, "\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)
	//
	//options := &password.Options{16, 100, 32, sha512.New}
	//salt, encodePwd := password.Encode("admin123", options)
	//newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodePwd)
	//fmt.Println(newPassword)

	//全局模式
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	host := fmt.Sprintf("http://192.168.1.4:9200")
	logger2 := log.New(os.Stdout, "ES", log.LstdFlags)
	global.EsClient, err = elastic.NewClient(elastic.SetURL(host), elastic.SetSniff(false), elastic.SetTraceLog(logger2))
	if err != nil {
		panic(err)
	}

	var goods []model.Goods
	db.Find(&goods)
	for _, g := range goods {
		esModel := model.EsGoods{
			ID:          g.ID,
			CategoryID:  g.CategoryID,
			BrandsID:    g.BrandsID,
			OnSale:      g.OnSale,
			ShipFree:    g.ShipFree,
			IsNew:       g.IsNew,
			IsHot:       g.IsHot,
			Name:        g.Name,
			ClickNum:    g.ClickNum,
			SoldNum:     g.SoldNum,
			FavNum:      g.FavNum,
			MarketPrice: g.MarketPrice,
			GoodsBrief:  g.GoodsBrief,
			ShopPrice:   g.ShopPrice,
		}

		_, err = global.EsClient.Index().Index(esModel.GetIndexName()).BodyJson(esModel).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
		if err != nil {
			panic(err)
		}
		//强调一下 一定要将docker启动es的java_ops的内存设置大一些 否则运行过程中会出现 bad request错误
	}
}
