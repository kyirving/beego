package utils

import (
	"fmt"
	"myBeego/components/redis"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

// 不需要验证登陆的路由
var NOTFILTERROUTER = []string{
	"/user/login",
	"/user/register",
}

type Filter struct {
	Response
}

func (f *Filter) FilterLoginStatus() beego.FilterFunc {
	filter := func(ctx *context.Context) {
		f.Ctx = ctx

		userId := f.GetString("userid")
		uri := ctx.Request.URL.Path
		exist := IsContainInList(NOTFILTERROUTER, uri)

		if userId == "" && !exist {

		}

		if !exist {

			//检查用户id是否上传
			if userId == "" {
				json := f.Error(RESP_SYSTEM_BUSY, "userid 不能为空")
				f.Ctx.Output.JSON(json, true, true)
				return
			}

			//检查token
			tokenInfo := f.Ctx.Request.Header["Access-Token"]
			if len(tokenInfo) == 0 {
				json := f.Error(RESP_PARAMS_ERROR, "Access-Token Can't be empty")
				f.Ctx.Output.JSON(json, true, true)
				return
			}

			token := tokenInfo[0]
			key := fmt.Sprintf("%s:user:tokenInfo:%s", beego.AppConfig.String("appname"), userId)
			accessToken, _ := redis.Rdb.HGet(key, "accessToken").Result()
			accessExprice, _ := redis.Rdb.HGet(key, "accessExprice").Result()

			if accessToken != token {
				json := f.Error(RESP_PARAMS_ERROR, "Access-Token invalid")
				f.Ctx.Output.JSON(json, true, true)
				return
			}

			access_exprice, _ := strconv.ParseInt(accessExprice, 10, 64)
			if access_exprice < time.Now().Unix() {
				json := f.Error(RESP_PARAMS_ERROR, "Access-Token 已过期请刷新Token")
				f.Ctx.Output.JSON(json, true, true)
				return
			}

		}

		// if f.Ctx.Output.Status != 0 {
		// 	f.Ctx.ResponseWriter.WriteHeader(f.Ctx.Output.Status)
		// } else {
		// 	f.Ctx.ResponseWriter.WriteHeader(500)
		// }

	}
	return filter
}

func IsContainInList(items []string, item string) bool {
	for _, v := range items {
		if item == v {
			return true
		}
	}
	return false
}
