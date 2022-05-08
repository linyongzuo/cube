package crackmodule

import (
	"context"
	"cube/config"
	"cube/gologger"
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"time"
)

type Mssql struct {
	*Crack
}

func (m *Mssql) CrackName() string {
	return "mssql"
}

func (m *Mssql) CrackPort() string {
	return "1433"
}

func (m *Mssql) CrackAuthUser() []string {
	return []string{"sa", "sql"}
}

func (m *Mssql) CrackAuthPass() []string {
	return config.PASSWORDS
}

func (m *Mssql) IsMutex() bool {
	return false
}

func (m *Mssql) CrackPortCheck() bool {
	return true
}

func (m *Mssql) Exec() CrackResult {
	//gologger.Debugf("exec begin")
	result := CrackResult{Crack: *m.Crack, Result: false, Err: nil}
	start := time.Now()
	//dataSourceName := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=master&connection+timeout=%d&encrypt=disable", m.Auth.User, m.Auth.Password, m.Ip,
	//	m.Port, m.Timeout)
	dataSourceName := fmt.Sprintf("odbc:server=%v;port=%v;user id=%v;database=%v;dial timeout=%d;encrypt=disable;connection timeout=%d;password={%s}", m.Ip,
		m.Port, m.Auth.User, "master", m.Timeout, m.Timeout, m.Auth.Password)
	db, err := sql.Open("mssql", dataSourceName)
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
						gologger.Infof("执行sql语句:%s 失败，数据库信息:%s:%s,错误原因:%s", v, m.Ip, m.Port, err.Error())
						continue
					} else {
						gologger.Infof("执行sql语句:%s 成功，数据库信息:%s:%s", v, m.Ip, m.Port)
					}
				}
			}
		} else {
			result.Extra = err.Error()
		}
	} else {
		result.Extra = err.Error()
	}
	result.CostTime = time.Now().Sub(start).Seconds()
	return result
}
func (m *Mssql) CrackMatch() (bool, string) {
	result := m.Exec()
	if !result.Result {
		return IsMssql(result.Extra), result.Extra
	}
	return result.Result, result.Extra
}
func init() {
	AddCrackKeys("mssql")
}
