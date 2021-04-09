package configs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"

	"cloudiac/consts"
)

type RedisConfig struct {
	IP       string `yaml:"ip"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type IamConfig struct {
	Addr    string `yaml:"addr"`
	AuthApi string `yaml:"authApi"`
}

type RabbitMqConfig struct {
	Addr  string `yaml:"addr"`
	Queue string `yaml:"queue"`
}

type yamlTimeDuration struct {
	time.Duration
}

type ConsulConfig struct {
	Address         string `yaml:"address"`
	ServiceID       string `yaml:"id"`
	ServiceIP       string `yaml:"ip"`
	ServicePort     int    `yaml:"port"`
	ServiceTags     string `yaml:"tags"`
	Interval        string `yaml:"interval"`
	Timeout         string `yaml:"timeout"`
	DeregisterAfter string `yaml:"deregister_after"`
}

type GitlabConfig struct {
	Url      string `yaml:"url"`
	Token    string `yaml:"token"`
	Username string `yaml:"username"`
}

type TaskConfig struct {
	TimeTicker int64  `yaml:"time_ticker"` // portal端轮训时间
	LogPath    string `yaml:"log_path"`
}

type RunnerConfig struct {
	AssetPath    string `yaml:"asset_path"`
	LogBasePaath string `yaml:"log_base_path"`
	DefaultImage string `yaml:"default_image"`
}

type LogConfig struct {
	LogMaxDays int    `yaml:"log_max_days"` // 日志文件保留天数, 默认 7
	LogLevel   string `yaml:"log_level"`
}

func (ut *yamlTimeDuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ds string
	if err := unmarshal(&ds); err != nil {
		return err
	}
	d, err := time.ParseDuration(ds)
	if err != nil {
		return err
	}
	ut.Duration = d
	return nil
}

type Config struct {
	Mysql                   string           `yaml:"mysql"`
	Redis                   RedisConfig      `yaml:"redis"`
	Listen                  string           `yaml:"listen"`
	Iam                     IamConfig        `yaml:"iam"`
	Rmq                     RabbitMqConfig   `yaml:"rabbitmq"`
	Prometheus              string           `yaml:"prometheus"`
	CollectTaskSyncInterval yamlTimeDuration `yaml:"collectTaskSyncInterval"`
	Consul                  ConsulConfig     `yaml:"consul"`
	Gitlab                  GitlabConfig     `yaml:"gitlab"`
	Runner                  RunnerConfig     `yaml:"runner"`
	Task                    TaskConfig       `yaml:"task"`
	Log                     LogConfig        `yaml:"log"`
}

var (
	gConfig *Config
	cfgLock sync.RWMutex
)

func parseConfig(filename string, out interface{}) error {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	bs = []byte(os.ExpandEnv(string(bs)))

	if err := yaml.Unmarshal(bs, out); err != nil {
		return fmt.Errorf("yaml.Unmarshal: %v", err)
	}
	return nil
}

func parsePortalConfig(filename string) error {
	cfg := Config{}
	if err := parseConfig(filename, &cfg); err != nil {
		return err
	}
	if cfg.CollectTaskSyncInterval.Duration == 0 {
		cfg.CollectTaskSyncInterval.Duration = consts.DefaultCollectTaskSyncInterval
	}

	cfgLock.Lock()
	defer cfgLock.Unlock()
	gConfig = &cfg

	return nil
}

func Get() *Config {
	cfgLock.RLock()
	defer cfgLock.RUnlock()

	return gConfig
}

func initConfig(filename string, parser func(string) error) {
	_, err := os.Stat(".env")
	if !os.IsNotExist(err) {
		if err := godotenv.Load(); err != nil {
			log.Panic(err)
		}
	}

	if err := parser(filename); err != nil {
		log.Panic(err)
	}
}

func Init(name string) {
	initConfig(name, parsePortalConfig)
}
