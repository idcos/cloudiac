package forms

import "cloudiac/portal/models"

type GcpDeployForm struct {
	BaseForm
	Region     string      `json:"region"`
	ZoneId     string      `json:"zoneId"`
	ChargeType string      `json:"chargeType"`
	ExtraData  models.JSON `form:"extraData" json:"extraData" binding:""` // 扩展字段，用于存储外部服务调用时的信息
}
