package util

var Void struct{}

var (
	TmpFilePath string = "/tmp/gate"
)

var (
	IpfsAddr string = ""
	RdsAddr  string = ""
	PgAddr   string = ""
)

type HttpConfig struct {
	Addr          string `json:"addr"`
	StaticSrcPath string `json:"static_src_path"`
}

func (c *HttpConfig) Check() {
	if c.Addr == "" {
		c.Addr = ":10086"
	}
}

type RecorderConfig struct {
	Capa int `json:"capa"`
}

func (c *RecorderConfig) Check() {
	if c.Capa <= 0 {
		c.Capa = 1
	}
}

type CacheConfig struct {
	Capa       int    `json:"capa"`
	FileExpire int    `json:"expire"` // 文件过期时间(单位:小时)
	Path       string `json:"cache_path"`
}

func (c *CacheConfig) Check() {
	if c.Capa <= 0 {
		c.Capa = 1
	}
	if c.FileExpire <= 0 {
		c.FileExpire = 1
	}
	if c.Path == "" {
		c.Path = GetCurrDir()
	}
}

type PgConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Passwd   string `json:"password"`
	DataBase string `json:"database"`
}

func (c *PgConfig) Check() {
	if c.Port < 0 || c.Port > 65535 {
		c.Port = 5432
	}
}

type RdsConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (c *RdsConfig) Check() {
	if c.Port < 0 || c.Port > 65535 {
		c.Port = 6379
	}
}

type IpfsConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	PullExpire int    `json:"pull_expire"` // 拉取超时(单位:分钟)
}

func (c *IpfsConfig) Check() {
	if c.Port < 0 || c.Port > 65535 {
		c.Port = 5001
	}
	if c.PullExpire <= 0 {
		c.PullExpire = 30
	}
}
