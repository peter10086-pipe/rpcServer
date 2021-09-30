package main

import (
	"fmt"
	ulog "github.com/ucloud/ucloud-sdk-go/ucloud/log"
	"log"
	"net/http"
	"net/rpc"
	"rpcServer/login"
	"sync"
	"time"
)

//go对RPC的支持，支持三个级别：TCP、HTTP、JSONRPC
//go的RPC只支持GO开发的服务器与客户端之间的交互，因为采用了gob编码

//注意字段必须是导出
type Params struct {
	SrcIp string;
	DstIp string;
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

    err:= login.U.VerifyLoginSuccess(p.Ips)

	if err != nil{

		ulog.Errorf("login fail ")
	}
	//fmt.Println("login.U.Clients",login.U.Clients)
	//var mux sync.WaitGroup
	//for i,ip1  := range p.Ips{
	//	for _ ,ip2:= range p.Ips[i+1:]{
	//		fmt.Println("ip2,ip1...",ip2,ip1,p.Ips[i:])
	//		time.Sleep(time.Millisecond*5)
	//		mux.Add(1)
	//		go func(ip3,ip4 string,cli map[string]*login.SSHClient){
	//			defer mux.Done()
	//			if cli == nil{
	//				return
	//			}
	//			raw := fmt.Sprintf("ping -c3 %s -I %s",ip3,ip4)
	//			std, err:= cli[ip4].SshSessionRun(raw)
	//			fmt.Println(std,err)
	//			if err!=nil{
	//				var a int
	//				a = -1
	//				ret = &a
	//			}
	//		}(ip2,ip1,login.U.Clients)
	//
	//	}
	//
	//}
	//mux.Wait()

	//var mux sync.WaitGroup
	//if true{
	//	for _, v := range p.Ips{
	//		mux.Add(1)
	//		go func(d string,ip Params){
	//			defer mux.Done()
	//			rawCmd := fmt.Sprintf("date;ping %s -c 1 -I %s",d,ip.SrcIp)
	//			fmt.Println("====================start",ip.SrcIp)
	//			fmt.Println(rawCmd)
	//			std,_:= login.SshHost(ip.SrcIp,rawCmd)
	//			fmt.Println("====================end")
	//			if strings.Contains(std,"100%"){
	//				var a int
	//				a = -1
	//				ret = &a
	//			}
	//			fmt.Println(std)
	//		}(v,p)
	//
	//	}
	//}

	//return nil;

	return nil
}

//func (r *VPC25Cube) Perimeter(p Params, ret *int) error {
//	*ret = (p.Width + p.Height) * 2;
//
//	return nil;
//}
func (r *VPC25Cube) ClientIperf(p Params, ret *int) error {

		fmt.Println(p)
		std, err := login.U.SshHost(p.SrcIp,"yum -y install iperf3")
		if err !=nil{
		return err
	}

		ulog.Infof("start yum -y install iperf3",std)

		log := time.Now().UnixNano()
		raw := fmt.Sprintf("nohup iperf3 -i2  -c %s -t20 > %v & sleep 20 ; cat %v",p.DstIp,log,log)
		std1, err := login.U.SshHost(p.SrcIp,raw)
		if err !=nil{
		return err
		}
		ulog.Infof(std1)
		return nil

}

func (r *VPC25Cube) Iperf(p Params, ret *int) error {

	fmt.Println(p)
	std, err := login.U.SshHost(p.SrcIp,"pkill iperf3;yum -y install iperf3")
	if err !=nil{
		return err
	}
	ulog.Infof("start p.SrcIp %s %s yum -y install iperf3",p.SrcIp,std)


	std1, err1 := login.U.SshHost(p.DstIp,"pkill iperf3;yum -y install iperf3")
	if err1 !=nil{
		return err1
	}

	ulog.Infof("start p.DstIp %s %s yum -y install iperf3",p.DstIp,std1)

	var sg sync.WaitGroup
	var sk sync.WaitGroup
	sg.Add(1)
	go func(){
		defer sg.Done()
		log := time.Now().UnixNano()
		raw := fmt.Sprintf("nohup timeout 30 iperf3 -i2 -s > %v & tail -f %v",log,log)
		std1, err := login.U.SshHost(p.DstIp,raw)
		if err !=nil{
			return
		}
		fmt.Println(std1)
		ulog.Infof(std1)
	}()
	time.Sleep(time.Second*3)
	sk.Add(1)
	go func(){
		defer sk.Done()
		log := time.Now().UnixNano()
		raw := fmt.Sprintf("nohup iperf3 -i2 -c %s -t20 > %v 2>&1 > %v and tail -f %v ",p.DstIp,log,log,log)
		std1, err := login.U.SshHost(p.SrcIp,raw)
		if err !=nil{
			fmt.Println("errerrerrerrerrerrerr",err)

			return
		}
		fmt.Println(std1)
		ulog.Infof(std1)
	}()

	sg.Wait()
	sk.Wait()
	fmt.Println("endendendendendendendendendendendend")

	return nil
}



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