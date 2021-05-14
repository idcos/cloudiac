package forms

type ConsulTagUpdateForm struct {
	BaseForm

	Tags      []string `json:"tags" form:"tags" `
	ServiceId string   `json:"serviceId" form:"serviceId" `
}
