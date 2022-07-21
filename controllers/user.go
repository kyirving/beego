package controllers

import (
	"fmt"
	"myBeego/components/redis"
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
	resp := &utils.Response{}

	this.ParseForm(loginParam)

	fmt.Println(loginParam)
	//定义用户模型
	user := models.User{}

	//查询单条记录
	err := o.QueryTable(user.TableName()).Filter("username", loginParam.Username).One(&user)
	if err == orm.ErrNoRows {
		this.Data["json"] = resp.Error(utils.RESP_PARAMS_ERROR, "用户名或密码错误(001)")
		this.ServeJSON()
		return
	}

	//密码校验
	if ok := models.CheckPasswordHash(loginParam.Password, user.PasswordHash); !ok {

		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "用户名或密码错误(002)")
		this.ServeJSON()
		return
	}

	if user.Status != models.STATUS_SUCC {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "账号异常")
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

	// n, err := utils.Rdb.Exists(key).Result()
	// if err != nil {
	// 	responseData.Code = utils.RESP_UNKNOW_ERROR
	// 	// responseData.Message = error.Error(err)
	// 	responseData.Message = error.Error(err)
	// 	this.Data["json"] = responseData
	// 	this.ServeJSON()
	// 	return
	// }

	// //先将该key删掉
	// if n > 0 {
	// 	redis.Rdb.Del(key)
	// }

	redis.Rdb.HSet(key, "accessToken", LoginResMsg.AccessToken).Result()
	redis.Rdb.HSet(key, "accessExprice", LoginResMsg.AccessExpire).Result()
	redis.Rdb.HSet(key, "refreshToken", LoginResMsg.RefreshToken).Result()
	redis.Rdb.HSet(key, "refreshExprice", LoginResMsg.RefreshExpire).Result()

	this.Data["json"] = resp.Success("登录成功", LoginResMsg)
	this.ServeJSON()
	return

}

func (this *UserController) Register() {
	o := orm.NewOrm()
	//this.ParseForm 用于解析到结构体

	resp := &utils.Response{}

	username := this.GetString("username")
	password := this.GetString("password")
	email := this.GetString("email")

	if username == "" {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "操作异常")
		this.ServeJSON()
		return
	}

	if password == "" {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "操作异常")
		this.ServeJSON()
		return
	}

	if email == "" {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "操作异常")
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
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "操作异常")
		this.ServeJSON()
		return
	}

	this.Data["json"] = resp.Success("注册成功")
	this.ServeJSON()
	return

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
	resp := &utils.Response{}
	_, err := qs.All(&users)
	if err != nil {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "操作异常")
		this.ServeJSON()
		return
	}

	this.Data["json"] = resp.Success("操作成功", users)
	this.ServeJSON()
}
