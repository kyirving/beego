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

type User struct {
	Id           int       `json:"id"`
	Username     string    `json:"userName" orm:"not null"`
	PasswordHash string    `json:"password_hash"`
	Email        string    `json:"email"`
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
