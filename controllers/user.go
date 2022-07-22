package controllers

import (
	"fmt"
	"myBeego/components/redis"
	"myBeego/components/utils"
	"myBeego/models"
	"os"
	"path"
	"strconv"
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
		Photo:        "static/img/default.jpeg",
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
	Page, _ := this.GetInt("page", 1)
	pageSize, _ := beego.AppConfig.Int("pageSize")
	pageSize, _ = this.GetInt("page_size", pageSize)

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

	count, _ := qs.Count()

	var users []*models.User
	resp := &utils.Response{}
	_, err := qs.Offset((Page - 1) * pageSize).Limit(pageSize).All(&users)
	if err != nil {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "操作异常")
		this.ServeJSON()
		return
	}

	result := make(map[string]interface{}, 2)
	pageInfo := &utils.PageInfo{
		Page:     Page,
		PageSize: pageSize,
		Total:    count,
	}
	result["pageInfo"] = pageInfo
	result["list"] = users

	this.Data["json"] = resp.Success("操作成功", result)
	this.ServeJSON()
}

/*
刷新token
*/
func (this *UserController) RefreshToken() {

	o := orm.NewOrm()
	resp := &utils.Response{}

	userId, _ := this.GetInt("userid")
	refresh_token := this.GetString("refresh_token")

	if userId == 0 {
		this.Data["json"] = resp.Error(utils.RESP_PARAMS_ERROR, "userid 不能为空!")
		this.ServeJSON()
		return
	}

	if refresh_token == "" {
		this.Data["json"] = resp.Error(utils.RESP_PARAMS_ERROR, "refresh_token 不能为空!")
		this.ServeJSON()
		return
	}

	user := models.User{
		Id: userId,
	}
	err := o.Read(&user)
	if err != nil {
		this.Data["json"] = resp.Error(utils.RESP_PARAMS_ERROR, "用户不存在")
		this.ServeJSON()
		return
	}

	key := fmt.Sprintf("%s:user:tokenInfo:%d", beego.AppConfig.String("appname"), userId)
	refreshToken, err := redis.Rdb.HGet(key, "refreshToken").Result()
	if err != nil || refreshToken != refresh_token {
		this.Data["json"] = resp.Error(utils.RESP_PARAMS_ERROR, "refreshToken异常，请重新登录")
		this.ServeJSON()
		return
	}

	refreshExprice, _ := redis.Rdb.HGet(key, "refreshExprice").Result()
	refresh_exprice, _ := strconv.ParseInt(refreshExprice, 10, 64)

	if refresh_exprice < time.Now().Unix() {
		this.Data["json"] = resp.Error(utils.RESP_PARAMS_ERROR, "refreshToken 已过期，请重新登录")
		this.ServeJSON()
		return
	}

	//生成token
	NewAccessToken := utils.RandToken(50)
	NewAccessExprice := time.Now().Unix() + 86400
	NewRefreshToken := utils.RandToken(50)
	NewRefreshExprice := time.Now().Unix() + (86400 * 7)
	LoginResMsg := models.LoginResMsg{
		UserId:        user.Id,
		AccessToken:   NewAccessToken,
		AccessExpire:  NewAccessExprice,
		RefreshToken:  NewRefreshToken,
		RefreshExpire: NewRefreshExprice,
	}

	redis.Rdb.HSet(key, "accessToken", LoginResMsg.AccessToken).Result()
	redis.Rdb.HSet(key, "accessExprice", LoginResMsg.AccessExpire).Result()
	redis.Rdb.HSet(key, "refreshToken", LoginResMsg.RefreshToken).Result()
	redis.Rdb.HSet(key, "refreshExprice", LoginResMsg.RefreshExpire).Result()

	this.Data["json"] = resp.Success("token刷新成功", LoginResMsg)
	this.ServeJSON()
	return
}

//更新头像
func (this *UserController) EditPhoto() {

	var (
		o    = orm.NewOrm()
		resp = &utils.Response{}
	)
	userId, _ := this.GetInt("userid")
	user := models.User{
		Id: userId,
	}
	err := o.Read(&user)
	if err != nil {
		this.Data["json"] = resp.Error(utils.RESP_NOT_FOUND, "用户不存在")
		this.ServeJSON()
		return
	}

	file, header, err := this.GetFile("file")
	if err != nil {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "文件获取失败")
		this.ServeJSON()
		return
	}
	defer file.Close()

	//文件格式判断
	fileExt := path.Ext(header.Filename)
	if fileExt != ".jpg" && fileExt != ".png" && fileExt != ".jpeg" {
		this.Data["json"] = resp.Error(utils.RESP_SUCC, "上传文件格式不正确")
		this.ServeJSON()
		return
	}

	time := time.Now()
	year := time.Year()
	month := time.Month()
	day := time.Day()
	directory := "static/img/photo/%d/%d/%d/"
	directory = fmt.Sprintf(directory, year, month, day)

	//目录不存在，则创建
	_, err = os.Stat(directory)
	if err != nil {
		if err = os.MkdirAll(directory, 0777); err != nil {
			beego.Error("os.MkdirAll err = ", err)
			this.Data["json"] = resp.Error(utils.RESP_SUCC, "创建目录失败")
			this.ServeJSON()
			return
		}
	}

	filename := fmt.Sprintf("%s%d%s", directory, time.Unix(), fileExt)

	err = this.SaveToFile("file", filename)
	if err != nil {
		beego.Error(err)
		this.Data["json"] = resp.Error(utils.RESP_SUCC, "上传文件失败")
		this.ServeJSON()
		return
	}

	//更新数据库
	user.Photo = filename
	if _, err := o.Update(&user); err != nil {

		beego.Error("o.Update err = ", err)

		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "修改失败")
		this.ServeJSON()
		return
	}

	this.Data["json"] = resp.Error(utils.RESP_SUCC, "修改头像成功")
	this.ServeJSON()
	return
}
