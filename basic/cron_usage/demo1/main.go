package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

/**
*
cronexpr很简单，
就是传入(* * * * * *)然后解析得到expr,
然后通过expr.Next()得到下次运行时间，
然后通过time.AfterFunc(nextTime.Sub(now))计算时间差
然后执行你的自己的任务
**/
func main()  {
	var (
		expr *cronexpr.Expression
		err	 error
		now  time.Time
		nextTime time.Time
	)

	//支持5位调度 分，时，天，月，星期
	if expr, err = cronexpr.Parse("* * * * *"); err != nil {
		fmt.Println(err)
	}

	//也支持7位调度 精确到秒、分、时、日、月、周、年
	if expr, err = cronexpr.Parse("*/5 * * * * * *"); err != nil {
		fmt.Println(err)
	}

	//当前时间
	now = time.Now()
	//下次调度时间
	nextTime = expr.Next(now)

	fmt.Println(now)
	fmt.Println(nextTime)

	//下次调度时间减去当前时间，然后回调
	time.AfterFunc(nextTime.Sub(now), func() {
		fmt.Println("被调度了", nextTime)
	})
	time.Sleep(time.Second * 5)
}