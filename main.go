package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"rpcServer/login"
	"strings"
	"sync"
)

//go对RPC的支持，支持三个级别：TCP、HTTP、JSONRPC
//go的RPC只支持GO开发的服务器与客户端之间的交互，因为采用了gob编码

//注意字段必须是导出
type Params struct {
	SrcIp string;
	Ips  []string
}



type VPC25Cube struct{}

//函数必须是导出的
//必须有两个导出类型参数
//第一个参数是接收参数
//第二个参数是返回给客户端参数，必须是指针类型
//函数还要有一个返回值error


func (r *VPC25Cube) FullMeshPing(p Params, ret *int) error {

	fmt.Println(p)

	var mux sync.WaitGroup
	if true{
		for _, v := range p.Ips{
			mux.Add(1)
			go func(d string,ip Params){
				defer mux.Done()
				rawCmd := fmt.Sprintf("date;ping %s -c 1 -I %s",d,ip.SrcIp)
				fmt.Println("====================start",ip.SrcIp)
				fmt.Println(rawCmd)
				std,_:= login.SshHost(ip.SrcIp,rawCmd)
				fmt.Println("====================end")
				if strings.Contains(std,"100%"){
					var a int
					a = -1
					ret = &a
				}
				fmt.Println(std)
			}(v,p)

		}
	}
	mux.Wait()
	return nil;
}

//func (r *VPC25Cube) Perimeter(p Params, ret *int) error {
//	*ret = (p.Width + p.Height) * 2;
//
//	return nil;
//}

func main() {
	VPC25Cube := new(VPC25Cube);
	//注册一个VPC25Cube服务
	rpc.Register(VPC25Cube);
	//把服务处理绑定到http协议上
	rpc.HandleHTTP();
	err := http.ListenAndServe("0.0.0.0:8082", nil);
	fmt.Println("start...")
	if err != nil {
		log.Fatal(err);
	}
}