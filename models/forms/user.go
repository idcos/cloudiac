package forms

type CreateUserForm struct {
	BaseForm
	UserName string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Phone    string `form:"phone" json:"phone" binding:""`
	Email    string `form:"email" json:"email" binding:""`

	TenantId uint `form:"tenantId" json:"tenantId"`
}

type UpdateUserForm struct {
	BaseForm
	Id          uint   `form:"id" json:"id" binding:""`
	UserName    string `form:"username" json:"username" binding:""`
	Phone       string `form:"phone" json:"phone" binding:""`
	//Email       string `form:"email" json:"email" binding:""`	// 邮箱不可编辑
	OldPassword string `form:"oldPassword" json:"oldPassword" binding:""`
	NewPassword string `form:"newPassword" json:"newPassword" binding:""`
}

type SearchUserForm struct {
	BaseForm

	Q          string `form:"q" json:"q" binding:""`
	Status     int    `form:"status" json:"status"`
	CustomerId uint   `form:"customerId" json:"customerId" binding:""`
}

type DeleteUserForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}

type DisableUserForm struct {
	BaseForm

	Id     uint `form:"id" json:"id" binding:"required"`
	Status int  `form:"status" json:"status" binding:"required"`
}

type DetailUserForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}

type InviteUserForm struct {
	BaseForm

	BaseURL   string   `json:"baseURL" form:"baseURL" binding:"required"`
	Emails    []string `json:"emails" form:"email" binding:"required"`
	Customers []uint   `json:"customers" form:"customer" binding:""`
	Roles     []uint   `json:"roles" form:"role"`
}

type ActivateUserForm struct {
	BaseForm

	Token    string `json:"token" form:"token" binding:"required"`
	Username string `json:"username" form:"username" binding:""`
	Password string `json:"password" form:"password" binding:""`
	Phone    string `json:"phone" form:"phone" binding:""`
}

type AdminSearchForm struct {
	BaseForm
}
