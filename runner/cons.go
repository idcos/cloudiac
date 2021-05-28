package runner


const (
	AssetPath = "/"
	ContainerLogPaht = "/var/run/"
	StaticFilePath = "/usr/yunji/cloudiac/tmp"
	DefaultImage = "mt5225/tf-ansible:v0.0.1"
	ContainerLogFilePath = "/usr/yunji/cloudiac/logs/"
	ContainerProviderPath = "/usr/yunji/cloudiac/provider"
	ContainerLogFileName = "runner.log"
	MaxLinesPreRead = 50
	ContainerEnvTerraform = "TF_PLUGIN_CACHE_DIR=/usr/yunji/cloudiac/provider"
	ContainerMountPath = "/usr/yunji/cloudiac"
	AnsibleStateAnalysis = "/usr/yunji/cloudiac/terraform.py"
)

//const ContainerKeysPath = "/usr/yunji/cloudiac/keys"
var (
	AnsibleEnv = map[string]string{
		"ANSIBLE_HOST_KEY_CHECKING":"False",
		"ANSIBLE_TF_DIR":"..",
		"ANSIBLE_NOCOWS":"1",
	}
)
