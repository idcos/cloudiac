package runner

import "time"

const ContainerLogFilePath = "/usr/yunji/cloudiac/logs/"

const ContainerProviderPath = "/usr/yunji/cloudiac/provider"

const ContainerLogFileName = "runner.log"

const MaxLinesPreRead = 50

const ContainerEnvTerraform = "TF_PLUGIN_CACHE_DIR=/usr/yunji/cloudiac/provider"

const ContainerMountPath = "/usr/yunji/cloudiac"

const FollowLogDelay = time.Second // follow 文件时读到 EOF 后进行下次读取的等待时长
