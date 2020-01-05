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
// 调度多个cron 任务

// 定义任务结果体
type CronJob struct {
	expr *cronexpr.Expression
	nextTime time.Time
}

func main() {
	// 需要一个 协程调度，定时检查所有Cron 任务，谁过期就执行谁
	var (
		cronJob *CronJob
		expr *cronexpr.Expression
		now time.Time
		scheduleTable map[string] *CronJob //key：任务名字
	)

	scheduleTable = make(map[string]*CronJob)

	// 当前时间
	now = time.Now()

	// 定义第一个Cronjob
	expr = cronexpr.MustParse("*/5 * * * * * *")
	cronJob = &CronJob{
		expr: expr,
		nextTime: expr.Next(now),
	}

	// 任务注册到调度表
	scheduleTable["job1"] = cronJob

	// 定义第二个cronjob
	expr = cronexpr.MustParse("*/8 * * * * * *")
	cronJob = &CronJob{
		expr : expr,
		nextTime: expr.Next(now),
	}
	// 任务注册到调度表
	scheduleTable["job2"] = cronJob

	// 启动调度协程
	go func() {
		var (
			jobName string
			cronJob *CronJob
			now time.Time
		)

		// 定时检查任务调度表是否有到期的
		for {
			now = time.Now()
			// 循环调度任务列表
			for jobName, cronJob = range scheduleTable {
				// 判断是否过期（如果下次调度时间等于当前时间，说明已经过期了）
				if cronJob.nextTime.Before(now) || cronJob.nextTime.Equal(now) {
					// 启动一个协程，执行这个任务
					go func(jobName string) {
						fmt.Println("执行：", jobName)
					}(jobName)
					// 计算下一次调度时间
					cronJob.nextTime = cronJob.expr.Next(now)
					fmt.Println(jobName, "下次执行时间：", cronJob.nextTime)
				}
			}
			// 睡眠100 毫秒（不让它占用过多cpu）
			select {
			case <- time.NewTimer(100 * time.Millisecond).C: //将在100 毫秒可读，返回
			}
		}
	}()
	time.Sleep(100 *time.Second)

}