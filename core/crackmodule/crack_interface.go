package crackmodule

import (
	"cube/gologger"
	"strings"
)

type Auth struct {
	User     string
	Password string
}
type Crack struct {
	Ip      string
	Port    string
	Auth    Auth
	Name    string
	Timeout int64
	Sql     []string
}

type CrackResult struct {
	Crack    Crack
	Result   bool
	Extra    string
	Err      error
	CostTime float64
}

var CrackKeys []string

type ICrack interface {
	CrackName() string       //插件名称
	CrackPort() string       //插件默认端口
	CrackAuthUser() []string //插件默认爆破的用户名
	CrackAuthPass() []string //插件默认爆破的密码，可以使用config.PASSWORD
	IsMutex() bool           //只能单独使用的插件，比如phpmyadmin
	CrackPortCheck() bool    //插件是否需要端口检查，一般TCP需要，phpmyadmin类单独使用的不用
	Exec() CrackResult       //运行插件
}

func AddCrackKeys(s string) {
	CrackKeys = append(CrackKeys, s)
}

func (c *Crack) NewICrack() ICrack {
	switch c.Name {
	case "ssh":
		return &SshCrack{c}
	case "ftp":
		return &FtpCrack{c}
	case "redis":
		return &Redis{c}
	case "elastic":
		return &Elastic{c}
	case "httpbasic":
		return &HttpBasic{c}
	case "jenkins":
		return &Jenkins{c}
	case "mongo":
		return &Mongo{c}
	case "mssql":
		return &Mssql{c}
	case "mysql":
		return &Mysql{c}
	case "postgres":
		return &Postgres{c}
	case "smb":
		return &Smb{c}
	case "zabbix":
		return &Zabbix{c}
	case "phpmyadmin":
		return &Phpmyadmin{c}
	case "oracle":
		return &Oracle{c}
	default:
		return nil
	}
}

func NewCrack(s string) Crack {
	return Crack{
		Name: s,
	}
}

func GetCrackPort(s string) string {
	c := NewCrack(s)
	ic := c.NewICrack()
	return ic.CrackPort()
}

func GetMutexStatus(s string) bool {
	c := NewCrack(s)
	ic := c.NewICrack()
	return ic.IsMutex()
}

func NeedPortCheck(s string) bool {
	c := NewCrack(s)
	ic := c.NewICrack()
	return ic.CrackPortCheck()
}

func NeedDatabaseCrack(addr IpAddr, timeout int64) (bool, string) {
	c := NewCrack(addr.PluginName)
	c.Port = addr.Port
	c.Ip = addr.Ip
	c.Auth.Password = "sa"
	c.Auth.User = "sa"
	c.Timeout = timeout
	ic := c.NewICrack()

	if ic.CrackName() == "mssql" {
		result := ic.Exec()
		if !result.Result {
			return IsMssql(result.Extra), result.Extra
		}
		return result.Result, result.Extra
	}
	return true, ""
}

func IsMssql(errMsg string) bool {
	return strings.Contains(errMsg, "login error")
}
func getPluginAuthUser(s string) []string {
	c := NewCrack(s)
	ic := c.NewICrack()
	return ic.CrackAuthUser()
}

func getPluginAuthPass(s string) []string {
	c := NewCrack(s)
	ic := c.NewICrack()
	return ic.CrackAuthPass()
}

func getPluginAuthCred(s string) bool {
	//检查插件是否设置了默认的用户和密码
	if len(getPluginAuthPass(s)) == 0 || len(getPluginAuthPass(s)) == 0 {
		return false
	}
	return true
}

func GetPluginAuthMap(s string) map[string][]Auth {
	auths := make([]Auth, 0)
	authMaps := make(map[string][]Auth, 0)
	credStatus := getPluginAuthCred(s)
	if !credStatus {
		gologger.Errorf("CrackAuthUser() or CrackAuthPass() is Empty for %s", s)
	}
	for _, user := range getPluginAuthUser(s) {
		for _, pass := range getPluginAuthPass(s) {
			pass = strings.Replace(pass, "{user}", user, -1)
			gologger.Debugf("%s is preparing default credentials: %s <=> %s", s, user, pass)
			auths = append(auths, Auth{
				User:     user,
				Password: pass,
			})
		}
	}
	authMaps[s] = auths
	return authMaps
}
