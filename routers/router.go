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
	beego.Router("/user/list", &controllers.UserController{}, "get:List")
	beego.Router("/user/refresh_token", &controllers.UserController{}, "post:RefreshToken")
	beego.Router("/user/edit_photo", &controllers.UserController{}, "post:EditPhoto")
}
