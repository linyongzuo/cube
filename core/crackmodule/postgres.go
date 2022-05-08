package crackmodule

import (
	"context"
	"cube/config"
	"cube/gologger"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type Postgres struct {
	*Crack
}

func (p *Postgres) CrackName() string {
	return "postgres"
}

func (p *Postgres) CrackPort() string {
	return "5432"
}

func (p *Postgres) CrackAuthUser() []string {
	return []string{"postgres", "admin", "root"}
}

func (p *Postgres) CrackAuthPass() []string {
	return config.PASSWORDS
}

func (p *Postgres) IsMutex() bool {
	return false
}

func (p *Postgres) CrackPortCheck() bool {
	return true
}

func (p *Postgres) Exec() CrackResult {
	result := CrackResult{Crack: *p.Crack, Result: false, Err: nil}

	dataSourceName := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v&connect_timeout=%d", p.Auth.User,
		p.Auth.Password, p.Ip, p.Port, "postgres", "disable", p.Timeout)
	db, err := sql.Open("postgres", dataSourceName)

	if err == nil {
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.Timeout)*time.Second)
		defer cancel()
		err = db.PingContext(ctx)
		if err == nil {
			result.Result = true
			if len(p.Sql) != 0 {
				for _, v := range p.Sql {
					_, err = db.ExecContext(ctx, v)
					if err != nil {
						gologger.Infof("插件类型:%s,执行sql语句:%s 失败，数据库信息:%s:%s,错误原因:%s", p.Name, v, p.Ip, p.Port, err.Error())
						continue
					} else {
						gologger.Infof("插件类型:%s,执行sql语句:%s 成功，数据库信息:%s:%s", p.Name, v, p.Ip, p.Port)
					}
				}
			}
		} else {
			result.Extra = err.Error()
		}
	}
	return result
}

func (p *Postgres) CrackMatch() (bool, string) {
	result := p.Exec()
	if !result.Result {
		return IsPostgres(result.Extra), result.Extra
	}
	return result.Result, result.Extra
}

func init() {
	AddCrackKeys("postgres")
}
