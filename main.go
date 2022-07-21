package main

import (
	"fmt"
	"myBeego/components"
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
	components.Rdb = components.ConnectRedisPool()
}

func main() {
	beego.Run("127.0.0.1:8080")
}
