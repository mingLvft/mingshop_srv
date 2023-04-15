package model

import (
	"context"
	"gorm.io/gorm"
	"mingshop_srvs/goods_srv/global"
	"strconv"
)

type Category struct {
	BaseModel
	Name             string      `gorm:"type:varchar(20);not null;comment:'分类名称'" json:"name"`
	ParentCategoryID int32       `gorm:"comment:'父分类ID'" json:"parent"`
	ParentCategory   *Category   `json:"-"`
	SubCategory      []*Category `gorm:"foreignKey:ParentCategoryID;references:ID" json:"sub_category"`
	Level            int32       `gorm:"type:int;not null;default:1;comment:'分类等级'" json:"level"`
	IsTab            bool        `gorm:"not null;default:false;comment:'是否导航栏'" json:"is_tab"`
}

type Brands struct {
	BaseModel
	Name string `gorm:"type:varchar(50);not null;comment:'品牌名称'"`
	Logo string `gorm:"type:varchar(200);not null;default:'';comment:'品牌logo'"`
}

type GoodsCategoryBand struct {
	BaseModel
	CategoryID int32 `gorm:"type:int;index:idx_category_brand,unique;comment:'分类ID'"`
	Category   Category
	BrandsID   int32 `gorm:"type:int;index:idx_category_brand,unique;comment:'品牌ID'"`
	Brands     Brands
}

func (GoodsCategoryBand) TableName() string {
	return "goodscategoryband"
}

type Banner struct {
	BaseModel
	Image string `gorm:"type:varchar(200);not null;comment:'轮播图'"`
	Url   string `gorm:"type:varchar(200);not null;comment:'轮播图跳转链接'"`
	Index int32  `gorm:"type:int;not null;default:1;comment:'轮播图顺序'"`
}

type Goods struct {
	BaseModel
	CategoryID      int32 `gorm:"type:int;comment:'分类ID'"`
	Category        Category
	BrandsID        int32 `gorm:"type:int;comment:'品牌ID'"`
	Brands          Brands
	OnSale          bool     `gorm:"not null;default:false;comment:'是否上架'"`
	ShipFree        bool     `gorm:"not null;default:false;comment:'是否包邮'"`
	IsNew           bool     `gorm:"not null;default:false;comment:'是否新品'"`
	IsHot           bool     `gorm:"not null;default:false;comment:'是否热销'"`
	Name            string   `gorm:"type:varchar(100);not null;comment:'品牌名称'"`
	GoodsSn         string   `gorm:"type:varchar(50);not null;comment:'商品唯一货号'"`
	ClickNum        int32    `gorm:"type:int;not null;default:0;comment:'点击数'"`
	SoldNum         int32    `gorm:"type:int;not null;default:0;comment:'商品销量'"`
	FavNum          int32    `gorm:"type:int;not null;default:0;comment:'收藏数'"`
	MarketPrice     float32  `gorm:"not null;comment:'市场价格'"`
	ShopPrice       float32  `gorm:"not null;comment:'本店价格'"`
	GoodsBrief      string   `gorm:"type:varchar(100);not null;comment:'商品简短描述'"`
	Images          GormList `gorm:"type:json;not null;comment:'商品图片'"`
	DescImages      GormList `gorm:"type:json;not null;comment:'商品详情图片'"`
	GoodsFrontImage string   `gorm:"type:varchar(200);not null;comment:'商品封面图'"`
}

func (g *Goods) AfterCreate(tx *gorm.DB) (err error) {
	esModel := EsGoods{
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
		return err
	}
	return nil
}
