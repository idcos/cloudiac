// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package configs

import (
	"crypto/md5" //nolint:gosec
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v2"
)

type KafkaConfig struct {
	Disabled     bool     `json:"disabled"`
	Brokers      []string `yaml:"brokers"`
	Topic        string   `yaml:"topic"`
	GroupID      string   `yaml:"group_id"`
	Partition    int      `yaml:"partition"`
	SaslUsername string   `yaml:"sasl_username"`
	SaslPassword string   `yaml:"sasl_password"`
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
	ConsulAcl       bool   `yaml:"consul_acl"`
	ConsulCertPath  string `yaml:"consul_cert_path"`
	ConsulAclToken  string `yaml:"consul_acl_token"`
	ConsulTls       bool   `yaml:"consul_tls"`
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

type LdapConfig struct {
	AdminDn          string `yaml:"admin_dn"`
	AdminPassword    string `yaml:"admin_password"`
	LdapServer       string `yaml:"ldap_server"`
	LdapServerPort   int    `yaml:"ldap_server_port"`
	SearchBase       string `yaml:"search_base"`
	SearchFilter     string `yaml:"search_filter"`   // 自定义搜索条件
	EmailAttribute   string `yaml:"email_attribute"` // 用户定义邮箱字段名 默认值为mail
	AccountAttribute string `yaml:"account_attribute"`
	OUSearchBase     string `yaml:"ou_search_base"`
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

type Config struct {
	Mysql              string           `yaml:"mysql"`
	Listen             string           `yaml:"listen"`
	Consul             ConsulConfig     `yaml:"consul"`
	Portal             PortalConfig     `yaml:"portal"`
	Runner             RunnerConfig     `yaml:"runner"`
	Log                LogConfig        `yaml:"log"`
	Kafka              KafkaConfig      `yaml:"kafka"`
	SMTPServer         SMTPServerConfig `yaml:"smtpServer"`
	SecretKey          string           `yaml:"secretKey"`
	JwtSecretKey       string           `yaml:"jwtSecretKey"`
	RegistryAddr       string           `yaml:"registryAddr"`
	ExportSecretKey    string           `yaml:"exportSecretKey"`
	HttpClientInsecure bool             `yaml:"httpClientInsecure"`
	Policy             PolicyConfig     `yaml:"policy"`
	Ldap               LdapConfig       `yaml:"ldap"`
	CostServe          string           `yaml:"cost_serve"`
	EnableTaskAbort    bool             `yaml:"enableTaskAbort"`
}

const (
	// 云模板等资源导出使用的默认加解密密钥，一般情况下不需要修改
	// 如果用户需要进一步加强安全性，可以通过配置 ExportSecretKey 环境变量修改该值
	defaultExportSecretKey = "rIhfbOpPsZHTdDA1yLJOxYNxCTFgTEuh" //nolint:gosec
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

	if err := ensureSecretKey(&cfg); err != nil {
		panic(err)
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

	if err := ensureSecretKey(&cfg); err != nil {
		panic(err)
	}

	lock.Lock()
	defer lock.Unlock()
	config = &cfg

	return nil
}

func ensureSecretKey(cfg *Config) error {
	if cfg.SecretKey == "" {
		return fmt.Errorf("missing secret key config")
	}
	// 如果 SecretKey 不是 32 位字符则使用 md5 转为 32 位
	if len(cfg.SecretKey) != 32 {
		hash := md5.New() //nolint:gosec
		hash.Write([]byte(cfg.SecretKey))
		cfg.SecretKey = fmt.Sprintf("%x", hash.Sum(nil))
	}
	return nil
}

// 直接传入 Config 结构设置 config
// 目前主要用于编写测试用例时直接初始化 config
func Set(cfg *Config) {
	config = cfg
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
