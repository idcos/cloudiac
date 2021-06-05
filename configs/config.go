package configs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type KafkaConfig struct {
	Brokers      []string `yaml:"brokers"`
	Topic        string   `yaml:"topic"`
	GroupID      string   `yaml:"group_id"`
	Partition    int      `yaml:"partition"`
	SaslUsername string   `yaml:"sasl_username"`
	SaslPassword string   `yaml:"sasl_password"`
	//TLSCertFile  string   `yaml:"tls_cert_file"`
	//TLSKeyFile   string   `yaml:"tls_key_file"`
	//TLSCAFile    string   `yaml:"tls_ca_file"`
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
	Type     string `yaml:"type"`
}

type TaskConfig struct {
	TimeTicker int64  `yaml:"time_ticker"` // portal端轮训时间
	LogPath    string `yaml:"log_path"`
}

type RunnerConfig struct {
	DefaultImage string `yaml:"default_image"`
	StoragePath  string `yaml:"storage_path"`
	ProviderPath string `yaml:"provider_path"`
}

type LogConfig struct {
	LogMaxDays int    `yaml:"log_max_days"` // 日志文件保留天数, 默认 7
	LogLevel   string `yaml:"log_level"`
}

type SMTPServerConfig struct {
	Addr     string `yaml:"addr"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`

	From     string `yaml:"from"`
	FromName string `yaml:"fromName"`
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
	Listen                  string           `yaml:"listen"`
	Consul                  ConsulConfig     `yaml:"consul"`
	Gitlab                  GitlabConfig     `yaml:"gitlab"`
	Runner                  RunnerConfig     `yaml:"runner"`
	Task                    TaskConfig       `yaml:"task"`
	Log                     LogConfig        `yaml:"log"`
	Kafka                   KafkaConfig      `yaml:"kafka"`
	SMTPServer              SMTPServerConfig `yaml:"smtpServer"`
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
	if err := parser(filename); err != nil {
		log.Panic(err)
	}
}

func Init(name string) {
	initConfig(name, parsePortalConfig)
}
