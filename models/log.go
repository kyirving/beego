package models

import (
	"encoding/json"

	"github.com/astaxie/beego/logs"
)

// beego 日志配置结构体
type LoggerConfig struct {
	FileName            string `json:"filename"` //将日志保存到的文件名及路径
	Level               int    `json:"level"`    // 日志保存的时候的级别，默认是 Trace 级别
	Maxlines            int    `json:"maxlines"` // 每个文件保存的最大行数，若文件超过maxlines，则将日志保存到下个文件中，为0表示不设置。默认值 1000000
	Maxsize             int    `json:"maxsize"`  // 每个文件保存的最大尺寸，若文件超过maxsize，则将日志保存到下个文件中，为0表示不设置。默认值是 1 << 28, //256 MB
	Daily               bool   `json:"daily"`    // 设置日志是否每天分割一次，默认是 true
	Maxdays             int    `json:"maxdays"`  // 设置保存最近几天的日志文件，超过天数的日志文件被删除，为0表示不设置，默认保存 7 天
	Rotate              bool   `json:"rotate"`   // 是否开启 logrotate，默认是 true
	Perm                string `json:"perm"`     // 日志文件权限
	RotatePerm          string `json:"rotateperm"`
	EnableFuncCallDepth bool   `json:"-"` // 输出文件名和行号
	LogFuncCallDepth    int    `json:"-"` // 函数调用层级
}

func LogsInit() {
	var logCfg = LoggerConfig{
		FileName:            "logs/beego.log",
		Level:               logs.LevelDebug,
		Daily:               true,
		EnableFuncCallDepth: true,
		LogFuncCallDepth:    3,
		RotatePerm:          "777",
		Perm:                "777",
	}

	// 设置beego log库的配置
	b, _ := json.Marshal(&logCfg)
	_ = logs.SetLogger(logs.AdapterFile, string(b))
	//logs.Async() //为了提升性能, 可以设置异步输出
	logs.Async(1e3) //异步输出允许设置缓冲 chan 的大小
}
