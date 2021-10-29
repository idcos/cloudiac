package forms

import "cloudiac/portal/models"

type CreateCiDeployForm struct {
	BaseForm
	ProjectName   string            `form:"projectName" json:"projectName" binding:"required"`
	EnvName       string            `form:"envName" json:"envName" binding:"required"`
	TplId         models.Id         `form:"tplId" json:"tplId" binding:"required"`
	TTL           string            `form:"ttl" json:"ttl" binding:"" enums:"0,12h,1d,3d,1w,15d,30d"` // 存活时间
	DestroyAt     string            `form:"destroyAt" json:"destroyAt" binding:""`                    // 自动销毁时间(时间戳)
	TfVarsFile    string            `form:"tfVarsFile" json:"tfVarsFile" binding:""`                  // Terraform tfvars 变量文件路径
	Playbook      string            `json:"playbook" form:"playbook"`
	KeyId         models.Id         `form:"keyId" json:"keyId" binding:""`                 // 部署密钥ID
	TerraformVars map[string]string `form:"terraformVars" json:"terraformVars" binding:""` // terraform 环境变量
	EnvVars       map[string]string `form:"envVars" json:"envVars" binding:""`             // 环境变量
}
