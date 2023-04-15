package initalize

import (
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
	"mingshop_srvs/userop_srv/global"
)

func InitRedisLock() {
	redisInfo := global.ServerConfig.RedisConfig
	client := goredislib.NewClient(&goredislib.Options{
		Addr: fmt.Sprintf("%s:%d", redisInfo.Host, redisInfo.Port),
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	global.RedisLock = redsync.New(pool)
}
