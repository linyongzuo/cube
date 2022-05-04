package crackmodule

import (
	"cube/config"
	"fmt"
	"net"
	"regexp"
	"strings"
)

type Redis struct {
	*Crack
}

func (r Redis) CrackName() string {
	return "redis"
}

func (r Redis) CrackPort() string {
	return "6379"
}

func (r Redis) CrackAuthUser() []string {
	return []string{""}
}

func (r Redis) CrackAuthPass() []string {
	return config.PASSWORDS
}

func (r Redis) IsMutex() bool {
	return false
}

func (r Redis) CrackPortCheck() bool {
	return true
}

func (r Redis) Exec() CrackResult {
	result := CrackResult{Crack: *r.Crack, Result: false, Err: nil}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", r.Ip, r.Port), config.TcpConnTimeout)
	if err != nil {
		return result
	}
	config, err := getConfig(conn)
	if err != nil {
		return result
	}

	if len(config) > 0 {
		result.Result = true
		result.Extra = fmt.Sprintf("Version=%s  OS=%s", config[0], config[1])
	} else {
		_, err = conn.Write([]byte(fmt.Sprintf("AUTH %s\r\n", r.Auth.Password)))
		if err != nil {
			return result
		}
		buf := make([]byte, 4096)
		count, _ := conn.Read(buf)
		response := string(buf[0:count])
		if strings.Contains(response, "+OK") {
			config, _ := getConfig(conn)
			result.Extra = fmt.Sprintf("Version=%s  OS=%s", config[0], config[1])

			result.Result = true
		}
	}
	return result
}
func (r Redis) CrackMatch() (bool, string) {
	return true, ""
}
func readReply(conn net.Conn) (result string, err error) {
	buf := make([]byte, 4096)
	for {
		count, err := conn.Read(buf)
		if err != nil {
			break
		}
		result += string(buf[0:count])
		if count < 4096 {
			break
		}
	}
	return result, err
}

func getConfig(conn net.Conn) (conf []string, err error) {
	_, err = conn.Write([]byte(fmt.Sprintf("INFO\r\n")))
	if err != nil {
		return
	}
	text, err := readReply(conn)
	if err != nil {
		return
	}
	if strings.Contains(text, "redis_version") {
		r := regexp.MustCompile(`.*redis_version:(.*)\n(?s).*(?U)os:(.*)\n`)

		match := r.FindStringSubmatch(text)
		a := strings.TrimSpace(match[1])
		b := strings.TrimSpace(match[2])
		conf = append(conf, a, b)
	}
	return conf, nil
}

func init() {
	AddCrackKeys("redis")
}
