package config

type GoodsSrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
	Port int    `mapstructure:"port" json:"port"`
}

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"db" json:"db"`
	User     string `mapstructure:"user" json:"user"`
	PassWord string `mapstructure:"password" json:"password"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type RedisConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type ServerConfig struct {
	Name        string       `mapstructure:"name" json:"name"`
	Tags        []string     `mapstructure:"tags" json:"tags"`
	Host        string       `mapstructure:"host" json:"host"`
	MysqlInfo   MysqlConfig  `mapstructure:"mysql" json:"mysql"`
	ConsulInfo  ConsulConfig `mapstructure:"consul" json:"consul"`
	RedisConfig RedisConfig  `mapstructure:"redis" json:"redis"`
	//商品微服务的配置
	GoodsSrvInfo GoodsSrvConfig `mapstructure:"goods_srv" json:"goods_srv"`
	//库存微服务的配置
	InventorySrvInfo GoodsSrvConfig `mapstructure:"inventory_srv" json:"inventory_srv"`
}

type NacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      uint64 `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	User      string `mapstructure:"user"`
	PassWord  string `mapstructure:"password"`
	DataId    string `mapstructure:"dataid"`
	Group     string `mapstructure:"group"`
}