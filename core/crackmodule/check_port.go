package crackmodule

import (
	"context"
	"cube/config"
	"cube/core"
	"cube/gologger"
	"fmt"
	"net"
	"sync"
	"time"
)

type IpAddr struct {
	Ip         string
	Port       string
	PluginName string
}

var (
	mutex        sync.Mutex
	checkMapLock sync.Mutex
	CheckMap     = map[string]IpAddr{}
	AliveAddr    = map[string]IpAddr{}
)

func CheckPort(ctx context.Context, threadNum int, delay float64, ips []string, pluginNames []string, port string, timeout int64) []IpAddr {
	//指定插件端口的时候，只允许加载一个插件
	var ipList []IpAddr
	if len(port) > 0 {
		for _, ip := range ips {
			//key := fmt.Sprintf("%s:%s:%s", ip, port, pluginNames[0])
			tmp := IpAddr{
				Ip:         ip,
				Port:       port,
				PluginName: pluginNames[0],
			}
			//checkMapLock.Lock()
			//if _, ok := CheckMap[key]; !ok {
			//	CheckMap[key] = tmp
			//	checkMapLock.Unlock()
			//} else {
			//	checkMapLock.Unlock()
			//	continue
			//}
			ipList = append(ipList, tmp)
		}
	} else {
		for _, plugin := range pluginNames {
			for _, ip := range ips {
				crackPort := GetCrackPort(plugin)
				tmp := IpAddr{
					Ip:         ip,
					Port:       crackPort,
					PluginName: plugin,
				}
				//checkMapLock.Lock()
				//key := fmt.Sprintf("%s:%s:%s", ip, crackPort, plugin)
				//if _, ok := CheckMap[key]; !ok {
				//	CheckMap[key] = tmp
				//	checkMapLock.Unlock()
				//} else {
				//	checkMapLock.Unlock()
				//	continue
				//}
				ipList = append(ipList, tmp)
			}
		}

	}
	if len(ipList) == 0 {
		return []IpAddr{}
	}
	var addrChan = make(chan IpAddr, len(ipList)*2)
	var wg sync.WaitGroup
	wg.Add(len(ipList))

	for i := 0; i < threadNum; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case addr, ok := <-addrChan:
					if !ok {
						return
					}
					if NeedPortCheck(addr.PluginName) || GetMutexStatus(addr.PluginName) {
						//TCP的时候是需要先端口检查,UDP跳过
						alive, msg := NeedDatabaseCrack(addr, timeout)
						if !alive {
							alive = false
							_, err := invalidFile.WriteString(fmt.Sprintf("错误的数据库IP地址:%s:%s,错误信息:%s\n", addr.Ip, addr.Port, msg))
							if err != nil {
								gologger.Warnf("写入文件失败:%s", err.Error())
							}
						} else {
							gologger.Infof("正确的数据库地址: %s:%s", addr.Ip, addr.Port)
						}
						SaveAddr(alive, addr)
					} else {
						gologger.Debugf("skip port check for %s", addr.PluginName)
						SaveAddr(true, addr)
					}
					wg.Done()
					select {
					case <-ctx.Done():
					case <-time.After(time.Duration(core.RandomDelay(delay)) * time.Second):
					}
				}
			}
		}()
	}

	for _, addr := range ipList {
		addrChan <- addr
	}
	close(addrChan)
	wg.Wait()

	return GetAliveAddr()
}
func CheckPortNew(ctx context.Context, threadNum int, delay float64, ipList []IpAddr, timeout int64) []IpAddr {
	//指定插件端口的时候，只允许加载一个插件

	if len(ipList) == 0 {
		return []IpAddr{}
	}
	var addrChan = make(chan IpAddr, len(ipList)*2)
	var wg sync.WaitGroup
	wg.Add(len(ipList))

	for i := 0; i < threadNum; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case addr, ok := <-addrChan:
					if !ok {
						return
					}
					if NeedPortCheck(addr.PluginName) || GetMutexStatus(addr.PluginName) {
						//TCP的时候是需要先端口检查,UDP跳过
						alive, msg := NeedDatabaseCrack(addr, timeout)
						if !alive {
							alive = false
							_, err := invalidFile.WriteString(fmt.Sprintf("错误的数据库IP地址:%s:%s,错误信息:%s\n", addr.Ip, addr.Port, msg))
							if err != nil {
								gologger.Warnf("写入文件失败:%s", err.Error())
							}
						} else {
							gologger.Infof("正确的数据库地址: %s:%s", addr.Ip, addr.Port)
						}
						SaveAddr(alive, addr)
					} else {
						gologger.Debugf("skip port check for %s", addr.PluginName)
						SaveAddr(true, addr)
					}
					wg.Done()
					select {
					case <-ctx.Done():
					case <-time.After(time.Duration(core.RandomDelay(delay)) * time.Second):
					}
				}
			}
		}()
	}

	for _, addr := range ipList {
		addrChan <- addr
	}
	close(addrChan)
	wg.Wait()

	return GetAliveAddr()
}
func check(addr IpAddr) (bool, IpAddr) {
	alive := false
	gologger.Debugf("port conn check: %s://%s:%s", addr.PluginName, addr.Ip, addr.Port)
	_, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", addr.Ip, addr.Port), config.TcpConnTimeout)
	if err == nil {
		gologger.Infof("Open %s:%s", addr.Ip, addr.Port)
		alive = true
		//conn.Close()
	}
	return alive, addr
}

//func checkUDP(addr IpAddr) (bool, IpAddr) {
//	//https://github.com/bronzdoc/gops
//	//alive := true
//	gologger.Debugf("skip udp port conn check: %s:%s", addr.Ip, addr.Port)
//	time.Sleep(time.Millisecond * 10)
//
//	return true, addr
//}

func SaveAddr(alive bool, addr IpAddr) {
	key := fmt.Sprintf("%s:%s:%s", addr.Ip, addr.Port, addr.PluginName)
	mutex.Lock()
	defer mutex.Unlock()
	if alive {
		AliveAddr[key] = addr
	}
}

func GetAliveAddr() []IpAddr {
	mutex.Lock()
	defer mutex.Unlock()
	tmp := make([]IpAddr, 0)
	for _, v := range AliveAddr {
		tmp = append(tmp, v)
	}
	AliveAddr = map[string]IpAddr{}
	return tmp
}
