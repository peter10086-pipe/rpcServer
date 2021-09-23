package login

import (
	"fmt"
	"net"
	"strconv"
	"time"
	"golang.org/x/crypto/ssh"
	"log"
)

const (

	UhostUsername      = "root"
	Password = "gauge_auto_test"
)

var (

	 globalClient map[string]*ssh.Session
)


func init(){

	globalClient = make(map[string]*ssh.Session)

}




func SshHost( host , rawCmd string) (stdout string, err error) {
	Commcli := NewSSHClient(host, UhostUsername, Password)
	stdout, err = Commcli.Run(rawCmd)
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
	var session *ssh.Session
	if _,ok := globalClient[c.IP];ok{
		fmt.Println("session has exist ",c.IP)
		session = globalClient[c.IP]

	}else{
	log.Println("execute cmd: %s,%s,%v", shell, c.Username+c.IP+strconv.Itoa(c.Port), t1)
	if c.Client == nil {
		if err := Retry(2, 1*time.Second, c.connect); err != nil {
			t2 := time.Now()
			log.Println("The connection failure took", t2.Sub(t1))
			return "", err
		}
	}
	t2 := time.Now()
	log.Println("The successful connection took  %v", t2.Sub(t1))
	var err1 error
	session, err1 = c.Client.NewSession()
	if err1 != nil {
		return "", err1
	}
	globalClient[c.IP] = session
	fmt.Println("globalClient",globalClient[c.IP])
	}
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

func (c *SSHClient) SshConnect() error {
	//Commcli := NewSSHClient(sshgw, vmId, sandboxSSHPassword, sandboxSSHPort)
	//defer c.client.Close()
	t1 := time.Now()
	if c.Client == nil {
		if err := Retry(20, 5*time.Second, c.connect); err != nil {
			t2 := time.Now()
			log.Println("The connection failure took  %v", t2.Sub(t1))
			return err
		}
	}
	t2 := time.Now()
	log.Println("The successful connection took  %v", t2.Sub(t1))
	//session, err := c.client.NewSession()
	//if err != nil {
	//	return ssh.Session{}, err
	//}
	//return *session, nil
	return nil
}

func (c *SSHClient) SshSessionRun(shell string) (string, error) {
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
