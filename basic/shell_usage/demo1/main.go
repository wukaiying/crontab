package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

/**
实现功能：
使用exec执行一个shell命令，sleep 2秒，在执行到1s的时候，中断协程，并把协成执行结果返回出来
 */
type result struct {
	outPut []byte
	err error
}

func main()  {

	var (
		ctx        context.Context
		cancelFunc context.CancelFunc
		cmd        *exec.Cmd
		resultChan chan *result																								//定chan负责协程和main主线程之间的数据传递
		res 	   *result
		)

	ctx, cancelFunc = context.WithCancel(context.TODO())																	//创建全局上下文

	resultChan = make(chan *result, 100)
	go func() {
		var (
			outPut 	[]byte
			err 	error
		)
		cmd = exec.CommandContext(ctx,"C://Program Files//Git//git-bash.exe","-c", "sleep 2; echo hello")							//在协成中执行exec需要使用该命令而不是exec.Command()
		if outPut, err = cmd.CombinedOutput(); err != nil {
			resultChan <- &result{																												//协程将数据传递到chan中
				outPut: outPut,
				err: err,
			}
		}
	}()

	//当主线程执行1秒的时候，执行cancelFunc中断协成
	time.Sleep(time.Second * 1)
	cancelFunc()																																	//结束协程
	res= <- resultChan
	fmt.Println(string(res.outPut))																													//main主线程接受到chan中的数
	fmt.Println(res.err)																															//main主线程接受到chan中的数
}
