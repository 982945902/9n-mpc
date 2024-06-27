package sdk

import (
	"bang/util"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func RegisterToProxy(redis_server string, redis_password string, id string, addr string) (err error) {
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

	host, port := util.ParseNodeAddr(addr)
	err = rdb.Set(ctx, fmt.Sprintf("network:%s", id), fmt.Sprintf("%s:%d", host, port), 0).Err()
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

var GlobalRegister struct {
	sc []constant.ServerConfig
	cc constant.ClientConfig
	c  naming_client.INamingClient
}

func OnceRegisterToNacos(id string, host string, nacosAddr string, metaData map[string]string) (err error) {
	nacosHost, nacosPort := util.ParserAddr(nacosAddr)
	GlobalRegister.sc = []constant.ServerConfig{
		*constant.NewServerConfig(nacosHost, uint64(nacosPort)),
	}
	GlobalRegister.cc = constant.ClientConfig{}

	GlobalRegister.c, err = clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &GlobalRegister.cc,
		ServerConfigs: GlobalRegister.sc,
	})
	if err != nil {
		return
	}

	host, port := util.ParseNodeAddr(host)
	success, err := GlobalRegister.c.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: "BanG",
		Ip:          host,
		Port:        uint64(port),
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   false,
		Metadata:    metaData,
	})
	if err != nil {
		return
	}

	if !success {
		err = fmt.Errorf("register to nacos failed")
		return
	}

	return
}
