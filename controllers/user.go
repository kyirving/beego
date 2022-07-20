package controllers

import (
	"myBeego/components"
	"myBeego/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) Login() {

	// this.ParseForm(obj interface{})
	userName := this.Ctx.Input.Param(":userName")
	userPwd := this.Ctx.Input.Param(":userPwd")

	this.Ctx.WriteString(userName)
	this.Ctx.WriteString(userPwd)
}

func (this *UserController) Register() {
	o := orm.NewOrm()
	//this.ParseForm 用于解析到结构体

	responseData := &components.ResponseHelper{}

	username := this.GetString("username")
	password := this.GetString("password")
	email := this.GetString("email")

	if username == "" {
		responseData.Code = components.RESP_PARAMS_ERROR
		responseData.Message = "用户名称不能为空"
		this.Data["json"] = responseData
		this.ServeJSON()
		return
	}

	if password == "" {
		responseData.Code = components.RESP_PARAMS_ERROR
		responseData.Message = "密码不能为空"
		this.Data["json"] = responseData
		this.ServeJSON()
		return
	}

	if email == "" {
		responseData.Code = components.RESP_PARAMS_ERROR
		responseData.Message = "邮箱不能为空"
		this.Data["json"] = responseData
		this.ServeJSON()
		return
	}

	passwordHash, _ := models.HashPassword(password)

	user := &models.User{
		Username:     username,
		PasswordHash: passwordHash,
		Email:        email,
		Status:       models.STATUS_SUCC,
		// Ctime:        time.Now().Format("2006-01-02 03:04:05"),
		// Uptime:       time.Now().Format("2006-01-02 03:04:05"),
	}

	_, err := o.Insert(user)
	if err != nil {
		responseData.Code = components.RESP_SYSTEM_BUSY
		responseData.Message = error.Error(err)

		this.Data["json"] = responseData
		this.ServeJSON()
		return
	}

	responseData.Code = components.RESP_SUCC
	responseData.Message = "注册成功"

	this.Data["json"] = responseData
	this.ServeJSON()

}
