package config

type Config struct {
	Name          string       `yaml:"Name"`
	Host          string       `yaml:"Host"`
	Port          int          `yaml:"Port"`
	MySQL         MySQLConf    `yaml:"MySQL"`
	Redis         RedisConf    `yaml:"Redis"`
	TDengine      TDengineConf `yaml:"TDengine"`
	EMQX          EMQXConf     `yaml:"EMQX"`
	JWTSecret     string       `yaml:"JWTSecret"`
	AdminPhone    string       `yaml:"AdminPhone"`
	AdminPassword string       `yaml:"AdminPassword"`
	CORSOrigins   string       `yaml:"CORSOrigins"` // 允许的域名，逗号分隔，* 表示全部
}

type MySQLConf struct {
	DSN string `yaml:"DSN"`
}

type RedisConf struct {
	Addr     string `yaml:"Addr"`
	Password string `yaml:"Password"`
	DB       int    `yaml:"DB"`
}

type TDengineConf struct {
	DSN string `yaml:"DSN"`
}

type EMQXConf struct {
	Broker   string `yaml:"Broker"`
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
	ClientID string `yaml:"ClientID"`
}
