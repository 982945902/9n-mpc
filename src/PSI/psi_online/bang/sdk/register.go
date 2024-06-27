package sdk

import (
	"context"
	"fmt"
	"net"

	"github.com/go-redis/redis/v8"
)

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func RegisterToProxy(redis_server string, redis_password string, id string, port int) (err error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_server,
		Password: redis_password,
		DB:       0,
	})

	ctx := context.Background()

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return
	}

	err = rdb.Set(ctx, fmt.Sprintf("network:%s", id), fmt.Sprintf("%s:%d", GetLocalIP(), port), 0).Err()
	if err != nil {
		return
	}

	return
}

func UnRegisterToProxy(redis_server string, redis_password string, id string) (err error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_server,
		Password: redis_password,
		DB:       0,
	})

	ctx := context.Background()

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return
	}

	err = rdb.Del(ctx, fmt.Sprintf("network:%s", id)).Err()
	if err != nil {
		return
	}

	return
}
