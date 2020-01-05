package worker

import (
	"fmt"
	"time"
	"wukaiying/crontab/crontab/common"
)

type Schedule struct {
	jobEventChan chan *common.JobEvent
	jobPlanTable map[string]*common.JobSchedulePlan
}

var (
	G_schedule *Schedule
)

func InitSchedule() (err error) {		//初始化chan和map,map和chan都是需要使用make来初始化的
	G_schedule = &Schedule{
		jobPlanTable: make(map[string] *common.JobSchedulePlan),
		jobEventChan: make(chan *common.JobEvent),
	}
	go G_schedule.scheduleLoop()
	return
}


func (schedule *Schedule) TrySchedule() (scheduleAfter time.Duration){
	//1.遍历所有任务
	var (
		jobPlan *common.JobSchedulePlan
		now time.Time
		nearTime *time.Time
	)
	now = time.Now()
	//如果没有任务则睡一秒，返回
	if len(schedule.jobPlanTable) == 0 {
		scheduleAfter = time.Second * 1
		return
	}

	for _, jobPlan = range schedule.jobPlanTable {				//遍历任务表
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {    //发现任务的nexttime小于等于当前时间，则需要执行了
			//jobPlanTable有任务开始执行
			fmt.Println("执行任务")
			fmt.Println("执行任务", jobPlan.Job.Name)
			jobPlan.NextTime = jobPlan.Expr.Next(now)			//更新下次执行时间
		}
		//2.统计最近一个将要过期的任务
		//3.更新调度时间间隔，更新调度时间间隔是为了不让for循环不停的执行，减少cpu消耗
		//计算一下距离最近下一次调度的时间和当前时间的间隔，sleep这个间隔，然后再执行
		//这里是计算时间间隔在scheduleloop中用
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	scheduleAfter = (*nearTime).Sub(now)
	return
}

//处理jobevent
func (schedule *Schedule) handleJobEvent(jobEvent *common.JobEvent)  {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		err error
		jobExisted bool
	)

	switch jobEvent.EventType {				//每一个jobevent具体处理逻辑，创建jobscheduleplan,分为Put和delete操作
	case common.JOB_EVENT_SAVE:
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		//然后将jobEventPlan加入到调度表里面
		schedule.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE:			//如果是删除操作则从map表中删除该key极其对应的值
		if jobSchedulePlan, jobExisted = schedule.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(schedule.jobPlanTable, jobEvent.Job.Name)
		}
	}
}

//协程
func (schedule *Schedule) scheduleLoop()  {
	var (
		jobEvent *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer			//定时调度器
	)

	//时间间隔
	scheduleAfter = schedule.TrySchedule()
	//定时调度器
	scheduleTimer = time.NewTimer(scheduleAfter)

	for {
		select {
		case jobEvent = <- schedule.jobEventChan:
			schedule.handleJobEvent(jobEvent)			//从eventchan中依次取出jobevent进行处理
		case <-scheduleTimer.C:							//定时任务到期
			//调度任务
			scheduleAfter = schedule.TrySchedule() 			//更新延时调度时间间隔
		}

		scheduleTimer.Reset(scheduleAfter)				//更新时间间隔
	}
}


//将jobevent放到jobenventchan中
func (schedule *Schedule) PushJobEvent(jobEvent *common.JobEvent) {
	schedule.jobEventChan <- jobEvent
}