package login

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
	cu "github.com/ucloud/ucloud-sdk-go/services/cube"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	ulog "github.com/ucloud/ucloud-sdk-go/ucloud/log"
	"golang.org/x/crypto/ssh"
	"sync"
	"time"
)

type (
	VpcfeClient struct {
		*ucloud.Client
	}
)

var U *UCloudEnv

func NewUCloudEnv() *UCloudEnv {
	config := ucloud.NewConfig()
	config.BaseUrl = "http://api.ucloud.cn"
	config.Region = "cn-sh2"
	config.Zone =  "cn-sh2-01"
	config.ProjectId ="org-0x4kng"

	if lvl, e := logrus.ParseLevel("debug"); e != nil {
		panic(e)
	} else {
		config.LogLevel = ulog.Level(lvl)
	}

	credential := auth.NewCredential()
	credential.PrivateKey = "EPToanhc560W5FzG1Zbq0QQK3h3kkf7hDOFyCv59SbCj68D9rOKp5sFzern9ULS5"
	credential.PublicKey = "gik0jB0CNWWgIbHrIr6ig3kIxrc0IoqTvu/huqf9u0ZRxA/8FEFUnxq7zOia8m2g"


	u := &UCloudEnv{
		Logger:       ulog.New(), // ulog.New(),
		cub:          cu.NewClient(&config,&credential),
		vpcfego:      NewVPCClient(&config,&credential),
		connects:     make(map[string]*ssh.Session),
		Clients:   make(map[string]*SSHClient),
	}
	return u
}

func NewVPCClient(config *ucloud.Config, credential *auth.Credential) *VpcfeClient {
	meta := ucloud.ClientMeta{Product: "VPC2.0"}
	client := ucloud.NewClientWithMeta(config, credential, meta)
	return &VpcfeClient{
		client,
	}
}

type  UCloudEnv struct {
	ulog.Logger
	cub            *cu.CubeClient
	vpcfego         *VpcfeClient
	connects         map[string]*ssh.Session
	Clients          map[string]*SSHClient

}


const (

	UhostUsername      = "root"
	Password = "gauge_auto_test"
)

var (

	 globalClient map[string]*ssh.Session
)


func init(){

	globalClient = make(map[string]*ssh.Session)
	U = NewUCloudEnv()
}

func (u *UCloudEnv) VerifyLoginSuccess(ips []string)error {
	var hostNames = make([]string, 0)
	type sshInfoSuccess struct {
		ip string
	}
	var successLoginHosts = make([]sshInfoSuccess, 0)
	//var InitClient = make(map[string]*SSHClient)
	//var mt sync.Mutex
	var wg sync.WaitGroup
	var errChan = make(chan error)

	//todo

	//

	for _, ip := range ips {
		wg.Add(1)
		go func( PodIp string) {
			defer wg.Done()
			//u.findHostIPByType(host.Name,"")
				u.Infof("current login host:%v", PodIp)
					cli := NewSSHClient(PodIp, UhostUsername, Password)
					if err := cli.SshConnect(); err != nil {
						//FailF(err, "%s(%s) login fail,other success login is %v", host.Name, host.UHostId, successLoginHosts)
						u.Errorf("%s(%s) login fail,other success login is %v", PodIp)
						errChan <- fmt.Errorf("internet err:%v,%s(%s) login fail, other success login is %v", err,successLoginHosts)
					} else {
						//cli.SshSessionRun(`echo "MaxSessions 1000000" >> /etc/ssh/sshd_config`)
						//cli.SshSessionRun(`echo "UseDNS no" >> /etc/ssh/sshd_config`)
						//cli.SshSessionRun(`systemctl restart sshd`)
						//cli.SshSessionRun(`arp -n|awk '/^[1-9]/{print "arp -d " $1}'|sh -x`)
						//cli.SshSessionRun(`ip a|grep  'inet .*/32' |sed  's/^\s*//'|sed 's/\s*$//'|awk '{print $2}' |xargs -I {} ip addr del {} dev eth0`)

						// cli.SshSessionRun(`echo 50000 > /proc/sys/net/ipv4/neigh/eth0/gc_stale_time`)
						//cli.Client.Close()
						//cli1 := NewSSHClient(PodIp, UhostUsername, Password)
						//if err := cli1.SshConnect(); err != nil {
						//	u.Errorf("%s(%s) login again fail,other success login is %v",PodIp)
						//	//FailF(err, "%s(%s) login fail,other success login is %v", host.Name, host.UHostId, successLoginHosts)
						//	errChan <- fmt.Errorf("login again internet err:%v,%s login fail, other success login is %v", err, PodIp, successLoginHosts)
						//} else {
						//	mt.Lock()
							u.Clients[PodIp] = cli
							a := ""
							log := time.Now().Unix()
						    for _,v:= range ips{
								a = a + fmt.Sprintf(" ping %s -c3 -I %s >> %v &",v,PodIp,log)
							}
						    a = a + fmt.Sprintf("sleep 5; cat %v;",log)
							//fmt.Sprintf("ping ")
							//cli.SshSessionRun(a)
						    std, err:=cli.SshSessionRun(a)

						    u.Infof("********************* src IP %v %v %v",PodIp,std,err)
							successLoginHosts = append(successLoginHosts, sshInfoSuccess{ip:PodIp})
							//mt.Unlock()
						//}
					}
		}(ip)
	}
	flg := false
	var reason error
	go func() {
		for {
			select {
			case errInfo := <-errChan:
				if errInfo != nil {
					flg = true
					reason = errInfo
				}
			default:

			}
		}
	}()

	wg.Wait()

	if flg {
		return fmt.Errorf("VerifyLoginSuccess error %s",flg)
	}
	u.Infof("hosts has login %v", ips)
	if len(ips) != 0 {
		u.Infof("子网%s主机%v已经确认能够全部登录", time.Now(),  hostNames)
	}
	return nil
}




func (u *UCloudEnv)  SshHost( host , rawCmd string) (stdout string, err error) {
	Commcli := NewSSHClient(host, UhostUsername, Password)
	stdout, err = Commcli.Run2(rawCmd)
	if err != nil {
		return fmt.Sprintf("ExecNormal fail %s", rawCmd), err
	}
	return
}


type SSHClient struct {
	IP         string
	Username   string
	Password   string
	Port       int
	Client     *ssh.Client
	LastResult string
}

//NewSSHClient 新建ssh 客户端
//@param ip IP地址
//@param username 用户名
//@param password 密码
//@param port 端口号,默认22
func NewSSHClient(ip string, username string, password string, port ...int) *SSHClient {
	cli := new(SSHClient)
	cli.IP = ip
	cli.Username = username
	cli.Password = password
	if len(port) <= 0 {
		cli.Port = 22
	} else {
		cli.Port = port[0]
	}
	return cli
}

//

//Run 执行shell命令并返回结果
//@param shell shell脚本命令
func (c *SSHClient) Run(shell string) (string, error) {
	t1 := time.Now()
	//var mtx sync.RWMutex
	//var session *ssh.Session
	//if _,ok := globalClient[c.IP];ok{
	//	fmt.Println("session has exist ",c.IP)
		//session = globalClient[c.IP]
	//
	//}else{
	log.Println("execute cmd ", shell, c.Username+c.IP+strconv.Itoa(c.Port), t1)
	if c.Client == nil {
		if err := Retry(10, 2*time.Second, c.connect); err != nil {
			t2 := time.Now()
			log.Println("The connection failure took", t2.Sub(t1))
			return "", err
		}
	}
	t2 := time.Now()
	log.Println("The successful connection took  %v", t2.Sub(t1))
	var err1 error
	session, err1 := c.Client.NewSession()
	if err1 != nil {
		return "", err1
	}
	//globalClient[c.IP] = session
	fmt.Println("globalClient",globalClient[c.IP])
	//}
	//defer session.Close()
	//mtx.Lock()
	//gauge.GetScenarioStore()[c.Username+c.IP+strconv.Itoa(c.Port)] = "true"
	//ConcurrentMap.Set(c.Username+c.IP+strconv.Itoa(c.Port), "true")
	//t.Sm.Map
	//mtx1.Unlock()
	buf, err := session.CombinedOutput(shell)
	c.LastResult = string(buf)
	//mtx1.Lock()
	//gauge.GetScenarioStore()[c.Username+c.IP+strconv.Itoa(c.Port)] = "false"
	//ConcurrentMap.Set(c.Username+c.IP+strconv.Itoa(c.Port), "false")
	//mtx.Unlock()
	return c.LastResult, err
}


func (c *SSHClient) Run2(shell string) (string, error) {
	t1 := time.Now()
	//var mtx sync.RWMutex
	//var session *ssh.Session
	//if _,ok := globalClient[c.IP];ok{
	//	fmt.Println("session has exist ",c.IP)
	//session = globalClient[c.IP]
	//
	//}else{
	log.Println("execute cmd ", shell, c.Username+c.IP+strconv.Itoa(c.Port), t1)
	if c.Client == nil {
		if err := Retry(10, 2*time.Second, c.connect); err != nil {
			t2 := time.Now()
			log.Println("The connection failure took", t2.Sub(t1))
			return "", err
		}
	}
	t2 := time.Now()
	log.Println("The successful connection took  %v", t2.Sub(t1))
	var err1 error
	session, err1 := c.Client.NewSession()
	if err1 != nil {
		return "", err1
	}
	//globalClient[c.IP] = session
	fmt.Println("globalClient",globalClient[c.IP])
	//}
	defer session.Close()
	//mtx.Lock()
	//gauge.GetScenarioStore()[c.Username+c.IP+strconv.Itoa(c.Port)] = "true"
	//ConcurrentMap.Set(c.Username+c.IP+strconv.Itoa(c.Port), "true")
	//t.Sm.Map
	//mtx1.Unlock()
	buf, err := session.CombinedOutput(shell)
	c.LastResult = string(buf)
	//mtx1.Lock()
	//gauge.GetScenarioStore()[c.Username+c.IP+strconv.Itoa(c.Port)] = "false"
	//ConcurrentMap.Set(c.Username+c.IP+strconv.Itoa(c.Port), "false")
	//mtx.Unlock()
	return c.LastResult, err
}


func (c *SSHClient) SshConnect() error {
	//Commcli := NewSSHClient(sshgw, vmId, sandboxSSHPassword, sandboxSSHPort)
	//defer c.client.Close()
	t1 := time.Now()
	if c.Client == nil {
		if err := Retry(10, 3*time.Second, c.connect); err != nil {
			t2 := time.Now()
			log.Println("The connection failure took ", t2.Sub(t1))
			return err
		}
	}
	t2 := time.Now()
	log.Println("The successful connection took ",c.IP, t2.Sub(t1))
	//session, err := c.client.NewSession()
	//if err != nil {
	//	return ssh.Session{}, err
	//}
	//return *session, nil
	return nil
}

func (c *SSHClient) SshSessionRun(shell string) (string, error) {
	fmt.Println(c.IP)
	session, err := c.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	log.Println(time.Now().String() + ":" + "start execute " + shell)
	buf, err := session.CombinedOutput(shell)
	log.Println(time.Now().String() + ":" + "end execute " + shell)
	return string(buf), err
}

//	defer session.Close()
//	buf, err := session.CombinedOutput(shell)
//
//	c.LastResult = string(buf)
//	return c.LastResult, err
//}
func noCheck(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}


//连接 远程连接目标ip
func (c *SSHClient) connect() error {
	config := ssh.ClientConfig{
		User:            c.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(c.Password)},
		HostKeyCallback: noCheck,
		Timeout:         60 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", c.IP, c.Port)
	sshClient, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return err
	}
	c.Client = sshClient
	return nil
}

//Retry 重试函数,规避主机连接失败长时间等待无返回报错退出，主机run起来的时候不一定能直接连接
func Retry(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if attempts--; attempts > 0 {
			fmt.Printf("retry func error: %s. attemps #%d after %s.", err.Error(), attempts, sleep)
			time.Sleep(sleep)
			return Retry(attempts, sleep, fn)
		}
		return err
	}
	return nil
}
