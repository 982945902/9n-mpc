package main

func main() {
	defer perface()()

	if globalConfig.Debug > 0 {
		debug()
		return
	}

	ctrl, err := newControler(&globalConfig)
	if err != nil {
		panic(err)
	}

	ctrl.Run()
}
