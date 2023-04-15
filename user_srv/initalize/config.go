package initalize

import (
	"fmt"
	"github.com/spf13/viper"
	"mingshop_srvs/user_srv/global"
)

func GetEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

func InitConfig() {
	//从配置文件中读取配置信息
	debug := GetEnvInfo("MINGSHOP_DEBUG")
	configFilePrefix := "config"
	configName := fmt.Sprintf("user_srv/%s-pro.yaml", configFilePrefix)
	if debug {
		configName = fmt.Sprintf("user_srv/%s-debug.yaml", configFilePrefix)
	}

	v := viper.New()
	v.SetConfigFile(configName)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(global.ServerConfig); err != nil {
		panic(err)
	}
}
