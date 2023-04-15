package model

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type GormList []string

func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), g)
}

type BaseModel struct {
	ID        int32      `gorm:"primary_key;type:int" json:"id"`
	CreatedAt time.Time  `gorm:"column:add_time;" json:"-"`
	UpdateAt  *time.Time `gorm:"column:update_time;" json:"-"`
	//UpdateAt  sql.NullTime     `gorm:"column:update_time;" json:"-"`
	DeleteAt  gorm.DeletedAt `gorm:"column:delete_time;" json:"-"`
	IsDeleted bool           `json:"-"`
}
