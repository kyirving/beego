package controllers

import (
	"context"
	"fmt"
	"io"
	"log"
	pd "myBeego/components/proto/myproto"
	"myBeego/components/utils"

	"github.com/astaxie/beego"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type TaskController struct {
	beego.Controller
}

func (this *TaskController) Exec() {

	commmand := "ls -la"
	address := "127.0.0.1:44445"

	// 客户端连接gRPC服务地址
	log.Printf("my-grpc Client grpc.Dial at Address:%s\n", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalln(err)
	}
	defer conn.Close()

	// 初始化客户端
	client := pd.NewCmdClient(conn)

	req := &pd.Request{}
	req.Method = "1"
	req.Command = commmand
	req.Spec = ""

	//调用方法 获取stream
	stream, err := client.ExecStream(context.Background(), req)
	if err != nil {
		log.Fatalf("could not echo: %v", err)
	}

	// for循环获取服务端推送的消息
	for {
		// 通过 Recv() 不断获取服务端send()推送的消息
		resp, err := stream.Recv()
		// err==io.EOF则表示服务端关闭stream了 退出
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Client Recv error:%v", err)
			continue
		}
		log.Printf("Client Recv data:%v\n", resp.GetMessage())
		fmt.Printf("Client Recv data:%v\n", resp.GetMessage())
	}

	resp := &utils.Response{}
	this.Data["json"] = resp.Success("请求成功")
	this.ServeJSON()
	return
}
