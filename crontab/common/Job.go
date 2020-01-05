package common

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

type Job struct {
	Name string `json:"name"`
	Command string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

type JobEvent struct {
	EventType int   // put delete
	Job *Job
}

type JobSchedulePlan struct {
	Job *Job						//需要调度的任务
	Expr *cronexpr.Expression		//任务调度表达式
	NextTime time.Time				//下次调度时间
}



//返回应答结构体
type Response struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}




func BuildResponse(code int, message string, data interface{}) (resp []byte, err error) {
	var (
		response Response
	)
	response.Code = code
	response.Message = message
	response.Data = data

	resp, err = json.Marshal(&response)
	return
}

func UnpackJob(content []byte) (job *Job, err error) {
	job = &Job{}
	if err = json.Unmarshal(content, job); err != nil {
		return
	}
	return job, nil
}

func ExtractJobName(value string) (str string) {
	str = strings.TrimPrefix(value, JOB_SAVE_DIR)
	return str
}

func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent)  {
	jobEvent = &JobEvent{
		EventType: eventType,
		Job: job,
	}
	return
}

func BuildJobSchedulePlan (job *Job) (jobSchedulePlan *JobSchedulePlan, err error)  {
	var (
		expr *cronexpr.Expression
	)

	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	jobSchedulePlan = &JobSchedulePlan{
		Job: job,
		Expr:expr,
		NextTime:expr.Next(time.Now()),
	}
	return
}