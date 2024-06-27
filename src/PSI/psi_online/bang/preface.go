package main

import (
	"bang/sdk"
	"fmt"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	flags "github.com/jessevdk/go-flags"
)

func perface() func() {
	_, err := flags.Parse(&globalConfig)
	if err != nil {
		panic(err)
	}

	report := sdk.NewStatusReport(globalConfig.StatusServer, globalConfig.Id, 0)

	report(-1, "", "", 0.0)

	return func() {
		r := recover()
		if r != nil {
			hlog.Errorf("panic:%v", r)
			report(500, fmt.Sprintf("%v", r), "", 0.0)
		} else {
			report(0, "", "", 0.0)
		}
	}
}
