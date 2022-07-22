package main

import (
	"context"
	"fmt"
	"myBeego/components/redis"
	"myBeego/components/utils"
	"myBeego/controllers"
	"myBeego/models"
	_ "myBeego/routers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

//初始化db：连接数据库
func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	// default必须要有,表示连接的数据库别名,可能是多个

	orm.Debug = true

	//注册默认数据库
	host := beego.AppConfig.String("db::host")
	port := beego.AppConfig.String("db::port")
	dbname := beego.AppConfig.String("db::databaseName")
	user := beego.AppConfig.String("db::userName")
	pwd := beego.AppConfig.String("db::password")

	dbcon := user + ":" + pwd + "@tcp(" + host + ":" + port + ")/" + dbname + "?charset=utf8"
	fmt.Println(dbcon)

	orm.RegisterDataBase("default", "mysql", dbcon)

	//初始化redis
	redis.Rdb = redis.ConnectRedisPool()
}

var Filter = func(ctx *context.Context) {
	fmt.Println(ctx)
}

func main() {

	// 路由过滤
	filter := &utils.Filter{}
	beego.InsertFilter("/user/*", beego.BeforeRouter, filter.FilterLoginStatus())

	//日志初始化
	models.LogsInit()

	//异常处理 todo
	beego.ErrorController(&controllers.ErrorController{})

	//注册样式：URL 前缀和映射的目录
	beego.Run("127.0.0.1:8080")

}
