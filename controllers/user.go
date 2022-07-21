package controllers

import (
	"fmt"
	"myBeego/components"
	"myBeego/components/utils"
	"myBeego/models"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) Login() {
	o := orm.NewOrm()

	var loginParam = &models.LoginParam{}
	responseData := &components.ResponseHelper{}

	this.ParseForm(loginParam)

	fmt.Println(loginParam)
	//定义用户模型
	user := models.User{}

	//查询单条记录
	err := o.QueryTable(user.TableName()).Filter("username", loginParam.Username).One(&user)
	if err == orm.ErrNoRows {
		responseData.Code = components.RESP_PARAMS_ERROR
		// responseData.Message = error.Error(err)
		responseData.Message = "用户名或密码错误(001)"
		this.Data["json"] = responseData
		this.ServeJSON()
		return
	}

	if user.Status != models.STATUS_SUCC {
		responseData.Code = components.RESP_SYSTEM_BUSY
		responseData.Message = "账号异常"
		this.Data["json"] = responseData
		this.ServeJSON()
		return
	}

	//密码校验
	if ok := models.CheckPasswordHash(loginParam.Password, user.PasswordHash); !ok {
		responseData.Code = components.RESP_PARAMS_ERROR
		// responseData.Message = error.Error(err)
		responseData.Message = "用户名或密码错误(002)"
		this.Data["json"] = responseData
		this.ServeJSON()
		return
	}

	//生成token
	accessToken := utils.RandToken(50)
	accessExprice := time.Now().Unix() + 86400
	refreshToken := utils.RandToken(50)
	refreshExprice := time.Now().Unix() + (86400 * 7)
	LoginResMsg := models.LoginResMsg{
		UserId:        user.Id,
		AccessToken:   accessToken,
		AccessExpire:  accessExprice,
		RefreshToken:  refreshToken,
		RefreshExpire: refreshExprice,
	}

	//将用户token存储在redis
	key := fmt.Sprintf("%s:user:tokenInfo:%d", beego.AppConfig.String("appname"), LoginResMsg.UserId)

	// n, err := components.Rdb.Exists(key).Result()
	// if err != nil {
	// 	responseData.Code = components.RESP_UNKNOW_ERROR
	// 	// responseData.Message = error.Error(err)
	// 	responseData.Message = error.Error(err)
	// 	this.Data["json"] = responseData
	// 	this.ServeJSON()
	// 	return
	// }

	// //先将该key删掉
	// if n > 0 {
	// 	components.Rdb.Del(key)
	// }

	components.Rdb.HSet(key, "accessToken", LoginResMsg.AccessToken).Result()
	components.Rdb.HSet(key, "accessExprice", LoginResMsg.AccessExpire).Result()
	components.Rdb.HSet(key, "refreshToken", LoginResMsg.RefreshToken).Result()
	components.Rdb.HSet(key, "refreshExprice", LoginResMsg.RefreshExpire).Result()

	responseData.Code = components.RESP_SUCC
	responseData.Message = "登录成功"
	responseData.Data = append(responseData.Data, LoginResMsg)
	this.Data["json"] = responseData
	this.ServeJSON()
	return

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

func (this *UserController) List() {
	UserName := this.GetString("username")
	Status, _ := this.GetInt("status")

	o := orm.NewOrm()

	//获取QuerySeter对象，user为表名
	user := &models.User{}
	qs := o.QueryTable(user.TableName())

	if UserName != "" {
		qs.Filter("username__contains", UserName)
	}

	if Status != 0 {
		qs.Filter("status", Status)
	}

	var users []*models.User
	var resp components.ResponseHelper
	_, err := qs.All(&users)
	if err != nil {
		resp.Code = components.RESP_SYSTEM_BUSY
		resp.Message = error.Error(err)

		this.Data["json"] = resp
		this.ServeJSON()
	}

	for _, v := range users {
		resp.Data = append(resp.Data, v)
	}

	resp.Code = components.RESP_SUCC
	resp.Message = "操作成功"

	this.Data["json"] = resp
	this.ServeJSON()

}
