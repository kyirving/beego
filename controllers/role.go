package controllers

import (
	"encoding/json"
	"fmt"
	"myBeego/components/utils"
	"myBeego/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type RoleController struct {
	beego.Controller
}

func (this *RoleController) Add() {
	var (
		resp = &utils.Response{}
		o    = orm.NewOrm()
	)
	var addParam = &models.AddParam{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &addParam)
	if err != nil {
		fmt.Println(this.Ctx.Input.RequestBody)
		this.Data["json"] = resp.Error(utils.RESP_PARAMS_ERROR, "获取参数失败")
		this.ServeJSON()
		return
	}

	role := &models.Role{
		Name: addParam.Name,
	}
	_, err = o.Insert(role)
	if err != nil {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "添加失败")
		this.ServeJSON()
		return
	}

	this.Data["json"] = resp.Success("添加成功")
	this.ServeJSON()
	return

}
