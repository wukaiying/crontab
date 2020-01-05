package worker

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"

	"wukaiying/crontab/crontab/common"
)

/**
将用户提交的数据存入etcd中，主要是conmmon/Job结构体
*/

type JobManager struct {
	client *clientv3.Client
	kv 		clientv3.KV
	lease 	clientv3.Lease
	watcher 	clientv3.Watcher
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
		watcher	clientv3.Watcher
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
	watcher = clientv3.Watcher(client)

	//赋值单利
	G_jobManager = &JobManager{
		client: client,
		kv:     kv,
		lease:  lease,
		watcher: watcher,
	}

	G_jobManager.watchJob()
	return
}

func (jobManager *JobManager) watchJob() (err error) {
	//1.获取所有的job,这是用户提交的job，我们需要拿到
	//2.获取当前reversion
	//3.根据当前reversion来watch后续任务的变化

	var (
		getResponse *clientv3.GetResponse
		job *common.Job
		watchStartRevision int64
		watchChan clientv3.WatchChan
		watchResponse clientv3.WatchResponse
		event *clientv3.Event
		jobName string
		jobEvent *common.JobEvent
	)
	if getResponse, err = jobManager.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}

	for _, kvpair := range getResponse.Kvs {
		//将json转化为结构体
		if job, err = common.UnpackJob(kvpair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			G_schedule.PushJobEvent(jobEvent)
		}
	}

	go func() {
		//从当前reversion监听事件变化，放到一个协程里面
		watchStartRevision = getResponse.Header.Revision + 1
		//监听/cron/jobs/后续变化
		watchChan = jobManager.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		for watchResponse = range watchChan {
			for _, event = range watchResponse.Events {
				switch event.Type {
				case mvccpb.PUT:
					if job, err = common.UnpackJob(event.Kv.Value); err != nil {
						continue
					}

					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
					fmt.Println(event.Kv.Value)
				case mvccpb.DELETE:
					jobName = common.ExtractJobName(string(event.Kv.Key))
					job = &common.Job{
						Name: jobName,
					}

					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
					fmt.Println(event.Kv.Value)
				}
				//send job event
				G_schedule.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

