package forms

type CreatePolicyForm struct {
	BaseForm

	Name          string `json:"name" binding:"required" example:"ECS分配公网IP"` // 策略名称
	ResourceType  string `json:"resourceType" binding:"required" example:"alicloud_instance"`
	FixSuggestion string `json:"fixSuggestion" binding:"" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"` // 修复建议
	Severity      string `json:"severity" binding:"" enums:"'high','medium','low','none'" example:"medium"`                            // 严重性

	PolicyType string `json:"policyType" binding:"" example:"alicloud"`
	Category   string `json:"category" binding:"" example:"basic"`

	Rego string `json:"rego" binding:"required"` // rego脚本
}

type CreatePolicyGroupForm struct {
	BaseForm

	Name        string `json:"name" binding:"required" example:"安全合规策略组"`
	Description string `json:"description" binding:"" example:"本组包含对于安全合规的检查策略"`

	PolicyIds []string `json:"policyIds" binding:"" example:"[\"po-c3ek0co6n88ldvq1n6ag\"]"`
}
