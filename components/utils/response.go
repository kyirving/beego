package utils

import "github.com/astaxie/beego"

const (
	RESP_SUCC                 = 200
	RESP_NOT_FOUND            = 404
	RESP_REQUEST_METHOD_ERROR = 405
	RESP_PARAMS_ERROR         = 406
	RESP_SYSTEM_BUSY          = 500
	RESP_ORDER_REPEAT         = 600
	RESP_UNKNOW_ERROR         = 601
	RESP_NETWORK_ERROR        = 602
)

type Response struct {
	beego.Controller
}

type any = interface{}

type Json struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (r *Response) Success(msg string, data ...any) Json {

	result := Json{Code: RESP_SUCC, Msg: msg}
	if len(data) > 0 {
		result.Data = data[0]
	}
	return result
}

func (r *Response) Error(code int, msg string) Json {
	result := Json{Code: code, Msg: msg}
	return result
}
