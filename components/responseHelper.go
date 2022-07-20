package components

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

type ResponseHelper struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    []interface{} `json:"data"`
}
