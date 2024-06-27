package main

import (
	"bang/mq/nats"
	"bang/sdk"
	"bang/util"
	"path/filepath"
)

func main() {
	defer perface()()

	if globalConfig.Debug > 0 {
		debug()
		return
	}

	globalConfig.StoragePath = filepath.Join(globalConfig.StoragePath, globalConfig.Id)

	ctrl, err := newControler(&globalConfig)
	if err != nil {
		panic(err)
	}

	metaData := map[string]string{}

	if globalConfig.EmbedMq {
		embedMqStoreDir := filepath.Join(globalConfig.StoragePath, "embed_mq")
		err := nats.StartNatsServer(globalConfig.MqAddress, embedMqStoreDir)
		if err != nil {
			panic(err)
		}
		metaData["embed-mq"] = "true"
		metaData["nacos-address"] = util.GetNodeAddr(globalConfig.NacosAddress)
	}

	err = sdk.OnceRegisterToNacos(globalConfig.Id, globalConfig.Host, globalConfig.NacosAddress, metaData)
	if err != nil {
		panic(err)
	}

	ctrl.Run()
}
