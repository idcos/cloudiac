package runner

import (
	"bytes"
	"cloudiac/utils"
	"os"
	"path/filepath"
	"text/template"
)

const cloudInitScriptTemplate = `#!/bin/sh
{{- if .PublicKey}}
mkdir -p /root/.ssh/ && \
echo '{{.PublicKey}}' >> /root/.ssh/authorized_keys && \
chmod 0600 /root/.ssh/authorized_keys
{{end -}}
`

/*
自动创建 backend 配置和 cloudinit 配置。

backend 用户不需要关心，会自动生效。

cloudinit 需要用户在所有计算资源中添加 user_data(各 provider 可能有不同) 属性。
以 aliyun 为例, 在 alicloud_instance 资源中添加 user_data 属性
并将值设置为 data.cloudinit_config.cloudiac.rendered

```hcl
resource "alicloud_instance" "instance" {
	// ...
	user_data = data.cloudinit_config.cloudiac.rendered
}
```

同时还会创建变量 cloudiac_private_key，值为 ssh 私钥路径
*/

const tfConfigTemplate = `terraform {
  backend "consul" {
    address = "{{.BackendAddress}}"
    scheme  = "{{.BackendScheme}}"
    path    = "{{.BackendPath}}"
    lock    = true
    gzip    = false
  }
}

data "cloudinit_config" "cloudiac" {
  gzip = false
  base64_encode = false

  part {
    content_type = "text/x-shellscript"
    content = <<EOT
{{.CloudInitContent}}
EOT
    filename = "_cloudiac-cloud-init.sh"
  }
}

variable "cloudiac_private_key" {
	default = "{{.ContainerTaskDir}}/ssh_key"
}
`

var (
	cloudInitTpl = template.Must(template.New("").Parse(cloudInitScriptTemplate))
	tfConfigTpl  = template.Must(template.New("").Parse(tfConfigTemplate))
)

type InjectConfigContext struct {
	WorkDir        string
	BackendAddress string
	BackendScheme  string
	BackendPath    string

	// 以下值自动设置，不需要传入
	CloudInitContent string
	ContainerTaskDir string
}

func GenInjectTfConfig(ctx InjectConfigContext, privateKey string) error {
	tplExecute := func(tpl *template.Template, savePath string, data interface{}) error {
		fp, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		defer fp.Close()
		return tpl.Execute(fp, data)
	}

	publicKey, err := utils.OpenSSHPublicKey([]byte(privateKey))
	if err != nil {
		return err
	}

	ctx.ContainerTaskDir = ContainerTaskDir
	// TODO: 改用 {{ template "name" }} 方式实现
	b := bytes.NewBuffer(nil)
	cloudInitTpl.Execute(b, map[string]interface{}{"PublicKey": string(publicKey)})
	ctx.CloudInitContent = b.String()

	if err := tplExecute(tfConfigTpl, filepath.Join(ctx.WorkDir, CloudIacTFName), ctx); err != nil {
		return err
	}
	os.WriteFile(filepath.Join(ctx.WorkDir, "ssh_key"), []byte(privateKey), 0600)
	return nil
}
