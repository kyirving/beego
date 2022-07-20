package routers

import (
	"myBeego/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	// beego.Router("/user/login", &controllers.UserController{})

	beego.Router("/user/login", &controllers.UserController{}, "post:Login")
	beego.Router("/user/register", &controllers.UserController{}, "post:Register")
}