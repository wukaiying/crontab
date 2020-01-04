package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"
	"wukaiying/crontab/crontab/master"
)

func initEnv()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var (
	configFilePath string
)
func initArgs()  {
	flag.StringVar(&configFilePath, "config", "./master.json", "input config file path")			//将用户传入的filepath赋值给configFilePath
}

func main()  {
	var (
		err error
	)
	initEnv()

	initArgs()

	if err = master.InitConfig(configFilePath); err != nil {
		goto ERR
	}


	//init job manager
	if err = master.InitJobManager(); err != nil {
		goto ERR
	}

	//inti api server
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	for{
		time.Sleep(time.Second)
	}

	return
	ERR:
		fmt.Println(err)

}

