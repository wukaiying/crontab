package master

import (
	"encoding/json"
	"io/ioutil"
)

/**
读取配置文件并赋给结构体
 */

type Config struct {
	ApiPort int `json:"apiPort"`
	ApiReadTimeout int `json:"apiReadTimeout"`
	ApiWriteTimeout int `json:"apiWriteTimeout"`
	EtcdEndPoints []string `json:"etcdEndPoints"`
	EtcdDailTimeout int `json:"etcdDailTimeout"`
	WebRoot string `json:"webRoot"`
}

var (
	G_config *Config
)

func InitConfig(filename string)  (err error){
	var (
		bytes []byte
		config Config
	)

	//从文件中读取config
	if bytes, err = ioutil.ReadFile(filename); err !=nil {
		return
	}
	//将json转化为结构体
	if err = json.Unmarshal(bytes,&config); err != nil {
		return
	}

	//赋值单利
	G_config = &config

	return
}
