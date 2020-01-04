package conmmon

import "encoding/json"

type Job struct {
	Name string `json:"name"`
	Command string `json:"command"`
	CronExpr string `json:"cronExpr"`
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