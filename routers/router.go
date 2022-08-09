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
	beego.Router("/user/change-status", &controllers.UserController{}, "post:ChangeStatus")
	beego.Router("/user/delete", &controllers.UserController{}, "delete:Delete")

	beego.Router("/role/add", &controllers.RoleController{}, "post:Add")

	//执行任务
	beego.Router("/task/exec", &controllers.TaskController{}, "post:Exec")
}
