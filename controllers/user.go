package controllers

import (
	"encoding/json"
	"fmt"
	"myBeego/components/redis"
	"myBeego/components/utils"
	"myBeego/models"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"

	"crypto/tls"

	"gopkg.in/gomail.v2"
)

type UserController struct {
	beego.Controller
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	wg = sync.WaitGroup{}
)

func (this *UserController) Login() {
	o := orm.NewOrm()

	var loginParam = &models.LoginParam{}
	data := this.Ctx.Input.RequestBody
	//json数据封装到user对象中
	err := json.Unmarshal(data, &loginParam)
	if err != nil {
		fmt.Println("json.Unmarshal is err:", err.Error())
	}
	resp := &utils.Response{}
	//定义用户模型
	user := models.User{}

	//查询单条记录
	err = o.QueryTable(user.TableName()).Filter("username", loginParam.Username).One(&user)
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
	var registerParam models.RegisterParam
	json.Unmarshal(this.Ctx.Input.RequestBody, &registerParam)

	resp := &utils.Response{}

	if registerParam.Username == "" {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "操作异常(1)")
		this.ServeJSON()
		return
	}

	if registerParam.Email == "" {
		this.Data["json"] = resp.Error(utils.RESP_SYSTEM_BUSY, "操作异常(3)")
		this.ServeJSON()
		return
	}

	//生产10位随机密码
	password := utils.RandToken(10)
	wg.Add(1)
	go SendMail(registerParam.Username, password)
	wg.Wait()

	d := gomail.NewDialer("smtp.example.com", 587, "user", "123456")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	passwordHash, _ := models.HashPassword(password)

	user := &models.User{
		Username:     registerParam.Username,
		PasswordHash: passwordHash,
		Email:        registerParam.Email,
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

	//发送邮件

	this.Data["json"] = resp.Success("注册成功")
	this.ServeJSON()
	return

}

func (this *UserController) List() {
	UserName := this.GetString("username")
	Status, _ := this.GetInt("status")
	sdate := this.GetString("sdate")
	edate := this.GetString("edate")
	Page, _ := this.GetInt("page", 1)
	pageSize, _ := beego.AppConfig.Int("pageSize")
	pageSize, _ = this.GetInt("page_size", pageSize)

	o := orm.NewOrm()

	//获取QuerySeter对象，user为表名
	user := &models.User{}
	qs := o.QueryTable(user.TableName())

	if UserName != "" {
		qs = qs.Filter("username__contains", UserName)
	}

	if Status != 0 {
		qs = qs.Filter("status", Status)
	}

	if sdate != "" && edate != "" {
		qs = qs.Filter("ctime__gte", sdate+" 00:00:00").Filter("ctime__lte", edate+" 23:59:59")
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

//应该放到utils 文件中
func SendMail(username string, pwd string) {
	message := `
    <p> Hello %s,</p>
	
		<p style="text-indent:2em">帐号已注册成功</p> 
		<p style="text-indent:2em">请到%s登录</p>
		<p style="text-indent:2em">初始化密码为:%s</P>
	`

	// QQ 邮箱：
	// SMTP 服务器地址：smtp.qq.com（SSL协议端口：465/994 | 非SSL协议端口：25）
	// 163 邮箱：
	// SMTP 服务器地址：smtp.163.com（端口：25）
	host := "smtp.163.com"
	port := 25
	userName := "15221478473@163.com"
	password := "QLFKGNJFXVHXJRYM" //授权码

	m := gomail.NewMessage()
	m.SetHeader("From", userName) // 发件人
	// m.SetHeader("From", "alias"+"<"+userName+">") // 增加发件人别名

	m.SetHeader("To", userName) // 收件人，可以多个收件人，但必须使用相同的 SMTP 连接
	// m.SetHeader("Cc", "******@qq.com")  // 抄送，可以多个
	// m.SetHeader("Bcc", "******@qq.com") // 暗送，可以多个
	m.SetHeader("Subject", "后台帐号注册成功") // 邮件主题

	// text/html 的意思是将文件的 content-type 设置为 text/html 的形式，浏览器在获取到这种文件时会自动调用html的解析器对文件进行相应的处理。
	// 可以通过 text/html 处理文本格式进行特殊处理，如换行、缩进、加粗等等
	m.SetBody("text/html", fmt.Sprintf(message, userName, "http://127.0.0.1:8081", pwd))

	// text/plain的意思是将文件设置为纯文本的形式，浏览器在获取到这种文件时并不会对其进行处理
	// m.SetBody("text/plain", "纯文本")
	// m.Attach("test.sh")   // 附件文件，可以是文件，照片，视频等等
	// m.Attach("lolcatVideo.mp4") // 视频
	// m.Attach("lolcat.jpg") // 照片

	d := gomail.NewDialer(
		host,
		port,
		userName,
		password,
	)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("d.DialAndSend err = ", err)
	}

	fmt.Println("sendmail success!!!")

	wg.Done()
}
