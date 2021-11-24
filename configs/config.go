// Copyright 2021 CloudJ Company Limited. All rights reserved.

package configs

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

type RunnerConfig struct {
	DefaultImage string `yaml:"default_image"`
	// AssetsPath  预置 providers 也在该目录下
	AssetsPath       string `yaml:"assets_path"`
	StoragePath      string `yaml:"storage_path"`
	PluginCachePath  string `yaml:"plugin_cache_path"`
	OfflineMode      bool   `yaml:"offline_mode"`       // 离线模式?
	ReserveContainer bool   `yaml:"reserver_container"` // 任务结束后保留容器?(停止容器但不删除)
}

type PortalConfig struct {
	Address       string `yaml:"address"` // portal 对外提供服务的 url
	SSHPrivateKey string `yaml:"ssh_private_key"`
	SSHPublicKey  string `yaml:"ssh_public_key"`
}

func (c *RunnerConfig) mustAbs(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return p
}

func (c *RunnerConfig) ProviderPath() string {
	if c.AssetsPath == "" {
		return ""
	}
	// 预置 providers 打包在 assets/providers 目录下，不单独提供配置
	return filepath.Join(c.AbsAssetsPath(), "providers")
}

func (c *RunnerConfig) AbsAssetsPath() string {
	return c.mustAbs(c.AssetsPath)
}

func (c *RunnerConfig) AbsStoragePath() string {
	return c.mustAbs(c.StoragePath)
}

func (c *RunnerConfig) AbsPluginCachePath() string {
	return c.mustAbs(c.PluginCachePath)
}

func (c *RunnerConfig) AbsTfenvVersionsCachePath() string {
	return c.mustAbs(filepath.Join(c.PluginCachePath, ".tfenv-versions"))
}

type LogConfig struct {
	LogLevel   string `yaml:"log_level"`
	LogPath    string `yaml:"log_path"`
	LogMaxDays int    `yaml:"log_max_days"` // 日志文件保留天数, 默认 7
}

type SMTPServerConfig struct {
	Addr     string `yaml:"addr"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`

	From     string `yaml:"from"`
	FromName string `yaml:"fromName"`
}

type PolicyConfig struct {
	Enabled bool `yaml:"enabled"`
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
	Mysql        string           `yaml:"mysql"`
	Listen       string           `yaml:"listen"`
	Consul       ConsulConfig     `yaml:"consul"`
	Portal       PortalConfig     `yaml:"portal"`
	Runner       RunnerConfig     `yaml:"runner"`
	Log          LogConfig        `yaml:"log"`
	Kafka        KafkaConfig      `yaml:"kafka"`
	SMTPServer   SMTPServerConfig `yaml:"smtpServer"`
	SecretKey    string           `yaml:"secretKey"`
	JwtSecretKey string           `yaml:"jwtSecretKey"`

	ExportSecretKey string `yaml:"exportSecretKey"`

	Policy PolicyConfig `yaml:"policy"`
}

const (
	defaultExportSecretKey = "rIhfbOpPsZHTdDA1yLJOxYNxCTFgTEuh"
)

var (
	config *Config
	lock   sync.RWMutex

	defaultConfig = Config{
		Portal: PortalConfig{
			SSHPrivateKey: "var/private_key",
			SSHPublicKey:  "var/private_key.pub",
		},
	}
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

func ParsePortalConfig(filename string) error {
	cfg := defaultConfig
	if err := parseConfig(filename, &cfg); err != nil {
		return err
	}
	if cfg.SecretKey == "" {
		panic("missing secret key config")
	}
	// 如果 SecretKey 不是 32 位字符则使用 md5 转为 32 位
	if len(cfg.SecretKey) != 32 {
		cfg.SecretKey = fmt.Sprintf("%x", md5.New().Sum([]byte(cfg.SecretKey)))
		fmt.Println(cfg.SecretKey ,"$$$$$$$", len(cfg.SecretKey))
	}

	if cfg.JwtSecretKey == "" {
		cfg.JwtSecretKey = cfg.SecretKey
	}
	if cfg.ExportSecretKey == "" {
		cfg.ExportSecretKey = defaultExportSecretKey
	}

	lock.Lock()
	defer lock.Unlock()
	config = &cfg

	return nil
}

func ParseRunnerConfig(filename string) error {
	cfg := defaultConfig
	if err := parseConfig(filename, &cfg); err != nil {
		return err
	}
	if cfg.SecretKey == "" {
		panic("missing secret key config")
	}
	lock.Lock()
	defer lock.Unlock()
	config = &cfg

	return nil
}

func Get() *Config {
	lock.RLock()
	defer lock.RUnlock()

	return config
}

func initConfig(filename string, parser func(string) error) {
	if err := parser(filename); err != nil {
		log.Panic(err)
	}
}

func Init(name string, parseFunc ...func(string) error) {
	if len(parseFunc) > 0 {
		initConfig(name, parseFunc[0])
	} else {
		initConfig(name, ParsePortalConfig)
	}
}
