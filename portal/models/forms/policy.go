package forms

type CreatePolicyForm struct {
	BaseForm

	//ReferenceId   models.Id `json:"reference_id" gorm:"not null;size:128;comment:策略ID" example:"iac_aliyun_public_26"`                                          // 策略ID
	Name          string `json:"name" gorm:"type:text;comment:名称" example:"ECS分配公网IP"`                                                                       // 策略名称
	FixSuggestion string `json:"fix_suggestion" gorm:"type:text;comment:策略修复建议" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"` // 修复建议
	Severity      string `json:"severity" gorm:"type:enum('high','medium','low','none');default:'medium';default:medium;comment:严重性" example:"medium"`       // 严重性

	PolicyType   string `json:"policy_type" gorm:"comment:云商类型" example:"alicloud"`
	ResourceType string `json:"resource_type" gorm:"comment:资源类型" example:"alicloud_instance"`
	Category     string `json:"category" gorm:"comment:分类"`

	Rego string `json:"rego" binding:"required"` // rego脚本
}
