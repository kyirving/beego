package models

import (
	"time"

	"github.com/astaxie/beego/orm"
	"golang.org/x/crypto/bcrypt"
)

const (
	STATUS_SUCC      = 1
	STATUS_ERR       = 2
	STATUS_EXCEPTION = 3
)

// 定义一个struct用来保存表单数据
// 通过给字段设置tag， 指定表单字段名， - 表示忽略这个字段不进行赋值
// 默认情况下表单字段名跟struct字段名同名（小写）
// type LoginParam struct {
// 	Username string `json:"username" form:"username"`
// 	Password string `json:"password" form:"password"`
// }

type LoginParam struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

type RegisterParam struct {
	Username string
	Email    string
}

type LoginResMsg struct {
	UserId        int    `json:"userId"`
	AccessToken   string `json:"accessToken"`
	AccessExpire  int64  `json:"accessExpire"`
	RefreshToken  string `json:"refreshToken"`
	RefreshExpire int64  `json:"refreshExpire"`
}

type User struct {
	Id           int       `json:"id"`
	Username     string    `json:"userName"`
	PasswordHash string    `json:"password_hash"`
	Email        string    `json:"email"`
	Photo        string    `json:"photo"`
	Status       int       `json:"status"`
	Ctime        time.Time `orm:"auto_now_add;type(timestamp)" json:"ctime"` //type(datetime)  auto_now_add:第一次保存时候的时间  auto_now:model保存时都会对时间自动更新
	Uptime       time.Time `json:"uptime" orm:"type(timestamp);auto_now_add;auto_now"`
}

//创建模型
func init() {
	orm.RegisterModel(new(User))
}

func (self *User) TableName() string {
	return "user"
}

//密码加密
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

//验证密码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
