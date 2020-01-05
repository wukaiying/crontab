package master

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
	"wukaiying/crontab/crontab/common"
)

type ApiServer struct {
	httpServer *http.Server
}

//单利模式
var (
	G_apiserver *ApiServer
)

//POST job={"name":"job1","command":"echo hello","cronExpr":"* * * * *"}
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var(
		err       error
		postJob   []byte
		job       common.Job
		oldJob    *common.Job
		respBytes []byte
	)
	//解析post表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	// 2. 取表单中的 job 字段
	//postJob = req.PostForm.Get("job")

	//获取用户Post数据
	if postJob, err = ioutil.ReadAll(req.Body); err != nil {
		goto ERR
	}

	//将用户传入转化为Job结构体
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}

	//将Job对象保存到etcd中
	if oldJob, err = G_jobManager.SaveJob(&job); err != nil {
		goto ERR
	}

	//返回应答给用户
	if respBytes, err = common.BuildResponse(0, "sava job success", oldJob); err == nil {
		resp.Write(respBytes)
	}
	return
ERR:
	if respBytes, err = common.BuildResponse(-1, err.Error(), oldJob); err == nil {
		resp.Write(respBytes)
	}
}


//POST /job/delete?name=xxx
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var(
		err error
		name string
		oldJob *common.Job
		respBytes []byte
	)
	//解析post表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	//获取用户Post数据
	name = req.FormValue("name")

	//将Job从etcd中删除
	if oldJob, err = G_jobManager.DeleteJob(name); err != nil {
		goto ERR
	}

	//返回应答给用户
	if respBytes, err = common.BuildResponse(0, "delete job success", oldJob); err == nil {
		resp.Write(respBytes)
	}
	return
ERR:
	if respBytes, err = common.BuildResponse(-1, err.Error(), oldJob); err == nil {
		resp.Write(respBytes)
	}
}

func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var(
		err error
		jobList []*common.Job
		respBytes []byte
	)

	//将Job从etcd中删除
	if jobList, err = G_jobManager.ListJob(); err != nil {
		goto ERR
	}

	//返回应答给用户
	if respBytes, err = common.BuildResponse(0, "list job success", jobList); err == nil {
		resp.Write(respBytes)
	}
	return
ERR:
	if respBytes, err = common.BuildResponse(-1, err.Error(), jobList); err == nil {
		resp.Write(respBytes)
	}
}

//POST /job/killer?name=xxx
//原理就是向/cron/killer下put一个key，worker节点监听这个key，然后在本地杀死进程
func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var(
		err error
		name string
		respBytes []byte
	)

	//解析表单
	if err = req.ParseForm(); err !=nil {
		goto ERR
	}

	//获取用户输入
	name = req.PostForm.Get("name")

	//将Job从etcd中删除
	if err = G_jobManager.KillJob(name); err != nil {
		goto ERR
	}

	//返回应答给用户
	if respBytes, err = common.BuildResponse(0, "kill job success", nil); err == nil {
		resp.Write(respBytes)
	}
	return
ERR:
	if respBytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(respBytes)
	}
}

func InitApiServer() (err error) {
	var (
		mux *http.ServeMux
		listener net.Listener
		httpServer *http.Server
		staticDir http.Dir 				//静态资源文件路径
		staticHandle http.Handler
	)

	//设置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete",handleJobDelete)
	mux.HandleFunc("/job/list",handleJobList)
	mux.HandleFunc("/job/kill",handleJobKill)

	staticDir = http.Dir(G_config.WebRoot)
	staticHandle = http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandle))

	//绑定端口
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		return
	}

	//创建http server
	httpServer = &http.Server{
		Handler:           mux,
		ReadTimeout:       time.Millisecond * time.Duration(G_config.ApiReadTimeout),
		WriteTimeout:      time.Millisecond * time.Duration(G_config.ApiWriteTimeout),
	}

	//启动server
	go httpServer.Serve(listener)

	G_apiserver = &ApiServer{
		httpServer:httpServer,
	}

	return
}