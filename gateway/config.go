package gateway

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/hanjingo/media_gateway/gateway/util"
)

var conf *Config
var confOnce = new(sync.Once)

func GetConf() *Config {
	confOnce.Do(func() {
		conf = &Config{
			Log:    util.LogConf(),
			Http:   &util.HttpConfig{},
			Record: &util.RecorderConfig{},
			Cache:  &util.CacheConfig{},
			Pg:     &util.PgConfig{},
			Rds:    &util.RdsConfig{},
			Ipfs:   &util.IpfsConfig{},
		}
	})
	return conf
}

type Config struct {
	Log    *util.LogConfig      `json:"log"`
	Http   *util.HttpConfig     `json:"http"`
	Record *util.RecorderConfig `json:"record"`
	Cache  *util.CacheConfig    `json:"cache"`
	Pg     *util.PgConfig       `json:"pg"`
	Rds    *util.RdsConfig      `json:"redis"`
	Ipfs   *util.IpfsConfig     `json:"ipfs"`
}

func (cfg *Config) Check() {
	cfg.Log.Check()
	cfg.Http.Check()
	cfg.Record.Check()
	cfg.Cache.Check()
	cfg.Pg.Check()
	cfg.Rds.Check()
	cfg.Ipfs.Check()
}

func (cfg *Config) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	d := json.NewDecoder(f)
	if err := d.Decode(cfg); err != nil {
		return err
	}
	cfg.Check()
	Log().Infof("load config from:%s success", filename)

	// set single var
	util.IpfsAddr = fmt.Sprintf("%s:%d", cfg.Ipfs.Host, cfg.Ipfs.Port)
	util.RdsAddr = fmt.Sprintf("%s:%d", cfg.Rds.Host, cfg.Rds.Port)
	// postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca
	// postgres://名字:密码@ip:端口/数据库名
	util.PgAddr = fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Pg.User, cfg.Pg.Passwd, cfg.Pg.Host, cfg.Pg.Port, cfg.Pg.DataBase)
	return nil
}
