package crackmodule

import (
	"bufio"
	"cube/config"
	"cube/gologger"
	"cube/pkg"
	"github.com/malfunkt/iprange"
	"os"
	"strconv"
	"strings"
)

type CrackOption struct {
	Ip         string
	IpFile     string
	User       string
	UserFile   string
	Pass       string
	PassFile   string
	Port       string
	PluginName string
	SqlFile    string
	Timeout    string
}

func NewCrackOptions() *CrackOption {
	return &CrackOption{}
}

func (cp *CrackOption) ParsePluginName() []string {
	var pluginNameList []string

	pns := strings.Split(cp.PluginName, ",")
	if len(pns) > 2 && pkg.Contains("X", pns) {
		//指定-X只能单独加载
		pluginNameList = nil
	}
	if len(pns) > 2 && pkg.Contains("Y", pns) {
		pluginNameList = nil
	}
	switch {
	case len(pns) == 1:
		if pns[0] == "X" {
			pluginNameList = config.CrackX
		}
		//if pns[0] == "Y" {
		//	for _, k := range CrackKeys {
		//		if !GetMutexStatus(k) {
		//			pluginNameList = append(pluginNameList, k)
		//		}
		//	}
		//}
		if pkg.Contains(pns[0], CrackKeys) {
			pluginNameList = pns
		}
	default:
		for _, k := range pns {
			if pkg.Contains(k, CrackKeys) {
				pluginNameList = append(pluginNameList, k)
			}
		}
	}
	return pluginNameList
}

func (cp *CrackOption) ParseAuth() []Auth {
	var auths []Auth
	user := cp.User
	userFile := cp.UserFile
	pass := cp.Pass
	passFile := cp.PassFile
	us := opt2slice(user, userFile, true)
	ps := opt2slice(pass, passFile, false)
	ps = append(ps, "")
	for _, u := range us {
		for _, p := range ps {
			auths = append(auths, Auth{
				User:     u,
				Password: p,
			})
		}
	}

	return auths
}

func (cp *CrackOption) ParseSql() []string {
	if len(cp.SqlFile) != 0 {
		return pkg.FileReader(cp.SqlFile, false)
	}
	return []string{}
}
func (cp *CrackOption) ParseIP() []string {
	var hosts []string
	ip := cp.Ip
	fp := cp.IpFile
	if ip != "" {
		hosts = ExpandIp(ip)
		if len(hosts) == 0 {
			os.Exit(1)
		}
	}

	if fp != "" {
		var ips []string
		ips, _ = ReadIPFile(fp)
		hosts = append(hosts, ips...)
	}
	hosts = pkg.RemoveDuplicate(hosts)
	return hosts
}

func (cp *CrackOption) ParseIpAddrFromFile() []IpAddr {
	var hosts []IpAddr
	fp := cp.IpFile
	if fp != "" {
		var ips []IpAddr
		ips, _ = ReadIPAddrFromFile(fp)
		hosts = append(hosts, ips...)
	}
	//hosts = pkg.RemoveDuplicate(hosts)
	return hosts
}
func (cp *CrackOption) ParsePort() bool {
	b, err := strconv.Atoi(cp.Port)
	if err != nil {
		gologger.Errorf("error while parse port option: %v", cp.Port)
	}
	if b < 0 || b > 65535 {
		gologger.Errorf("incorrect port: %v", cp.Port)
		return false
	}
	return true
}
func (cp *CrackOption) GetTimeout() string {
	return cp.Timeout
}
func opt2slice(str, file string, trimSpace bool) []string {
	if len(str+file) == 0 {
		gologger.Errorf("Provide login name(-l/-L) and login password(-p/-P)")
	}
	if len(str) > 0 {
		r := strings.Split(str, ",")

		return r
	}
	r := pkg.FileReader(file, trimSpace)
	return r
}

func ExpandIp(ip string) (hosts []string) {

	list, err := iprange.ParseList(ip)
	if err != nil {
		gologger.Warnf("IP parsing error\nformat: 10.0.0.1, 10.0.0.5-10, 192.168.1.*, 192.168.10.0/24")
		return
	}
	rng := list.Expand()
	for _, v := range rng {
		hosts = append(hosts, v.String())

	}
	return hosts

}

func ReadIPFile(filename string) ([]string, error) {
	file, err := os.OpenFile(filename, os.O_RDWR, 0766)
	if err != nil {
		gologger.Debugf("Open %s error, %s\n", filename, err)
	}
	defer file.Close()
	var content []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			all := strings.Split(text, " ")
			ip := ""
			if len(all) == 1 {
				ip = strings.TrimSpace(all[0])
			} else {
				if len(all) >= 4 {
					ip = all[3]
				} else {
					continue
				}
			}
			host := ExpandIp(ip)
			if len(host) == 0 {
				continue
			}
			content = append(content, host...)
		}
	}
	// 清空文件
	err = file.Truncate(0)
	_, err = file.Seek(0, 0)
	return content, nil
}

func ReadIPAddrFromFile(filename string) ([]IpAddr, error) {
	file, err := os.OpenFile(filename, os.O_RDWR, 0766)
	if err != nil {
		gologger.Debugf("Open %s error, %s\n", filename, err)
	}
	defer file.Close()
	var content []IpAddr
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			all := strings.Split(text, " ")
			ip := ""
			port := ""
			if len(all) == 1 {
				ip = strings.TrimSpace(all[0])
			} else {
				if len(all) >= 4 {
					ip = all[3]
					port = all[2]
				} else {
					continue
				}
			}
			name := ""
			switch port {
			case "1433":
				name = "mssql"
			case "5631":
				name = "postgres"
			default:
				continue
			}
			content = append(content, IpAddr{
				Ip:         ip,
				Port:       port,
				PluginName: name,
			})
		}
	}
	// 清空文件
	err = file.Truncate(0)
	_, err = file.Seek(0, 0)
	return content, nil
}
