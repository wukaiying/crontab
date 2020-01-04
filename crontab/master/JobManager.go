package master

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"go.etcd.io/etcd/clientv3"
	"time"
	"wukaiying/crontab/crontab/conmmon"
)

/**
将用户提交的数据存入etcd中，主要是conmmon/Job结构体
 */

type JobManager struct {
	client *clientv3.Client
	kv 		clientv3.KV
	lease 	clientv3.Lease
}

//共外界访问
var (
	G_jobManager *JobManager
)

//初始化连接etcd
func InitJobManager() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv 		clientv3.KV
		lease   clientv3.Lease
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints: G_config.EtcdEndPoints,
		DialTimeout: time.Millisecond * time.Duration(G_config.EtcdDailTimeout),
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	//得到kv
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	//赋值单利
	G_jobManager = &JobManager{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

//save Job to etcd,保存格式为/cron/jobs/名字 -> Job结构体的 json格式
//如果是覆盖则返回被覆盖的job,新增的话就正常添加
func (jobManager *JobManager)SaveJob(job *conmmon.Job) (oldJob *conmmon.Job, err error)  {

	var(
		jobKey string
		jobValue []byte
		putResponse *clientv3.PutResponse
		oldJobObj conmmon.Job
	)

	jobKey = conmmon.JOB_SAVE_DIR + job.Name
	if jobValue, err = json.Marshal(job); err!= nil {
		return
	}

	fmt.Println("save job is:", &job)
	glog.V(3).Info("save job is:", &job)
	//保存到etcd
	if putResponse, err = jobManager.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}

	glog.V(3).Info("job value is :", jobValue)
	fmt.Println("job value is :", jobValue)

	//如果是覆盖，则会返回旧值
	if putResponse.PrevKv != nil {
		if err = json.Unmarshal(putResponse.PrevKv.Value, &oldJobObj); err != nil {
			return
		}
		oldJob = &oldJobObj									//给返回值变量赋值，这样可以简化return
	}
	return
}


func (jobManager *JobManager)DeleteJob(name string) (oldJob *conmmon.Job, err error)  {
	var (
		jobKey string
		deleteResp *clientv3.DeleteResponse
		oldJobObj conmmon.Job
	)

	//删除etcd中的key
	jobKey = conmmon.JOB_SAVE_DIR + name
	if deleteResp, err = jobManager.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	//返回被删除的信息
	if len(deleteResp.PrevKvs) != 0 {
		if err = json.Unmarshal(deleteResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			return
		}
		oldJob = &oldJobObj
	}
	return
}

func (jobManager *JobManager)ListJob() (jobList []*conmmon.Job, err error)  {
	var (
		keyDir string
		listResp *clientv3.GetResponse
		jobObj *conmmon.Job
		jobObjList []*conmmon.Job
	)

	//list 以keyDir为前缀的所有job
	keyDir = conmmon.JOB_SAVE_DIR
	if listResp, err = jobManager.kv.Get(context.TODO(), keyDir, clientv3.WithPrefix()); err != nil {
		return
	}
	//返回查询到的
	jobObjList = make([]*conmmon.Job, 0)							//这样保证，如果没有Job返回的结果不会为空，而是一个空数组
	for _, item := range listResp.Kvs {
		jobObj = &conmmon.Job{}
		if err = json.Unmarshal(item.Value, jobObj); err != nil {
			err = nil										//容忍这次错误
			continue
		}
		jobObjList = append(jobObjList, jobObj)
	}

	jobList = jobObjList

	return
}

//save Job to etcd,保存格式为/cron/kill/名字 -> ""
//作用仅仅是触发一次put操作
func (jobManager *JobManager)KillJob(name string) (err error) {

	var(
		jobKey string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseID clientv3.LeaseID
	)

	jobKey = conmmon.JOB_KILL_DIR + name

	//创建租约
	if leaseGrantResp,err = jobManager.lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	//获取lease id
	leaseID = leaseGrantResp.ID

	//保存到etcd
	if _, err = jobManager.kv.Put(context.TODO(), jobKey, "", clientv3.WithLease(leaseID)); err != nil {
		return
	}
	return
}