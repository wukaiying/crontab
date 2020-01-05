package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"
	"wukaiying/crontab/crontab/worker"
)

func initEnv()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var (
	configFilePath string
)
func initArgs()  {
	flag.StringVar(&configFilePath, "config", "./worker.json", "input config file path")			//将用户传入的filepath赋值给configFilePath
}

func main()  {
	var (
		err error
	)
	initEnv()

	initArgs()

	//init config
	if err = worker.InitConfig(configFilePath); err != nil {
		goto ERR
	}

	if err = worker.InitSchedule(); err != nil {
		goto ERR
	}

	//init job manager
	if err = worker.InitJobManager(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(time.Second)
	}
	

	return
ERR:
	fmt.Println(err)

}
