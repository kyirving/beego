package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

type AddParam struct {
	Name string
}

type Role struct {
	Id     int       `json:"role_id" orm:"column(user_id)"`
	Name   string    `json:"name" orm:"unique"`
	Isdel  int       `json:"isdel" orm:"default(1)"`
	Ctime  time.Time `orm:"auto_now_add;type(datetime)" json:"ctime"` //type(datetime)  auto_now_add:第一次保存时候的时间  auto_now:model保存时都会对时间自动更新
	Uptime time.Time `orm:"auto_now;type(datetime)" json:"uptime"`
}

//创建模型
func init() {
	orm.RegisterModel(new(Role))
}

func (self *Role) TableName() string {
	return "role"
}
