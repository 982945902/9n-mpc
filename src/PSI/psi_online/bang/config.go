package main

type Config struct {
	Debug int `long:"debug" default:"0" description:"debug mode"`

	Id     string `long:"application-id" description:"application id"`
	Domain string `long:"domain" description:"serve domain"`
	Host   string `long:"host" description:"serve bind host"`

	Target    string `long:"target" description:"peer domain"`
	Remote    string `long:"remote" description:"peer host"`
	StorePath string `long:"store-path" description:"storage path"`

	RedisServer   string `long:"redis-server" description:"redis server"`
	RedisPassword string `long:"redis-password" description:"redis password"`

	StatusServer string `long:"status-server" description:"status server"`

	NodeID int `long:"node-id" description:"node id"`

	MqAddress    string `long:"mq-address" description:"mq address"`
	NacosAddress string `long:"nacos-address" description:"nacos address"`

	LinkHost string `long:"link-host" description:"serve bind host"`

	WindowSize     int  `long:"window-size" default:"1" description:"send window size"`
	// RecoverSupport bool `long:"recover-support" default:"false" description:"if server support recover"`
}

var globalConfig = Config{}
