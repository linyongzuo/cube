package crackmodule

import (
	"context"
	"cube/config"
	"cube/gologger"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type Mysql struct {
	*Crack
}

func (m *Mysql) CrackName() string {
	return "mysql"
}

func (m *Mysql) CrackPort() string {
	return "3306"
}

func (m *Mysql) CrackAuthUser() []string {
	return []string{"root", "mysql"}
}

func (m *Mysql) CrackAuthPass() []string {
	return config.PASSWORDS
}

func (m *Mysql) IsMutex() bool {
	return false
}

func (m *Mysql) CrackPortCheck() bool {
	return true
}

func (m *Mysql) Exec() CrackResult {
	start := time.Now()
	result := CrackResult{Crack: *m.Crack, Result: false, Err: nil}

	dataSourceName := fmt.Sprintf("%v:%v@tcp(%v:%v)/mysql?charset=utf8&timeout=%v", m.Auth.User, m.Auth.Password, m.Ip, m.Port, m.Timeout)
	db, err := sql.Open("mysql", dataSourceName)
	if err == nil {
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.Timeout)*time.Second)
		defer cancel()
		err = db.PingContext(ctx)
		if err == nil {
			result.Result = true
			if len(m.Sql) != 0 {
				for _, v := range m.Sql {
					_, err = db.ExecContext(ctx, v)
					if err != nil {
						gologger.Infof("执行sql语句:%s 失败，数据库信息:%v,错误原因:%s", v, *m, err.Error())
						break
					} else {
						gologger.Infof("执行sql语句:%s 成功，数据库信息:%v", v, *m)
					}
				}
			}
		}
		db.Close()
	}
	result.CostTime = time.Now().Sub(start).Seconds()
	return result
}
func (m *Mysql) CrackMatch() (bool, string) {
	result := m.Exec()
	if !result.Result {
		return IsMssql(result.Extra), result.Extra
	}
	return result.Result, result.Extra
}
func init() {
	AddCrackKeys("mysql")
}
