package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

/**
多个定时任务调度
将所有的定时任务存储到map中，然后一个协程去遍历map判断定时任务是否需要执行
https://blog.csdn.net/u011327801/article/details/90376402
 */

//定义任务
type CronJob struct {
	expr 			*cronexpr.Expression					//描述定时任务执行策略
	nextTime		time.Time
}

func main(){
	var (
		cronJob *CronJob
		now		time.Time
		expr 	*cronexpr.Expression
		err		error
		scheduleTable		map[string]*CronJob						//定时任务注册表
	)
	now = time.Now()
	scheduleTable = make(map[string]*CronJob)						//map需要使用make事先进行创建
	//定义第一个任务
	if expr, err = cronexpr.Parse("*/5 * * * * * *"); err != nil {
		fmt.Println(err)
	}
	cronJob = &CronJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	//注册任务到scheduleTable
	scheduleTable["job1"] = cronJob

	//定义第二个任务
	if expr, err = cronexpr.Parse("*/8 * * * * * *"); err != nil {
		fmt.Println(err)
	}
	cronJob = &CronJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	scheduleTable["job2"] = cronJob

	//启动协程遍历map执行定时调度任务
	go func() {
		var (
			now time.Time
			jobName string
			jobCron *CronJob
		)
		for {
			now = time.Now()
			for jobName, jobCron = range scheduleTable {
				if jobCron.nextTime.Before(now) && jobCron.nextTime.Equal(now) {					//下次调度时间已经小于等于当前时间，说明需要进行调度
					go func(jobName string) {
						fmt.Println("正在执行：" + jobName)														//打印jobName
					}(jobName)

					//还要更新下次调度时间
					jobCron.nextTime = jobCron.expr.Next(now)
					fmt.Println(jobName, "下次执行时间：", cronJob.nextTime)
				}
			}

			//为for 循环添加一个延时时间，每100ms执行一次，等价于time.sleep(100*time.Millisecond)
			select {
			case <- time.NewTimer(100 * time.Millisecond).C:
			}
		}
	}()

	time.Sleep(100 *time.Second)								//让主线程等待




}