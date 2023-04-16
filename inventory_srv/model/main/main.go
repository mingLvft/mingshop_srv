package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"mingshop_srvs/inventory_srv/model"
	"os"
	"time"
)

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/mingshop_inventory_srv?charset=utf8mb4&parseTime=True&loc=Local"

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

	//for i := 0; i < 10; i++ {
	//	user := model.User{
	//		NickName: fmt.Sprintf("ming%d", i),
	//		Mobile:   fmt.Sprintf("1333333333%d", i),
	//		Password: newPassword,
	//	}
	//	db.Save(&user)
	//}

	//
	////定义一个表结构，将表结构直接生成对应的表 -migrations
	////迁移 schema
	_ = db.AutoMigrate(&model.Inventory{}, &model.StockSellDetail{})
}
