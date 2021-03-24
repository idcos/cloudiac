package e

const (
	// 通用
	InternalError           = 1000
	ObjectAlreadyExists     = 1010
	ObjectNotExists         = 1011
	ObjectNotExistsOrNoPerm = 1013 // 对象不存在或者无权限
	ObjectDisabled          = 1014
	JSONParseError          = 1015
	StateTryAgainLater      = 1016
	ColValidateError        = 1017

	NotImplement = 1020

	DBError           = 1030 // db 操作出错
	DBAttrValidateErr = 1031

	ToFromAttrErr = 1032

	QuotaError = 1033

	BadParam               = 1040 // 参数错误(参数值不对)
	BadRequest             = 1041 // 请求错误(请求缺少必要参数)
	IOError                = 1050 // 文件 io 出错
	LdapError              = 1060 // ldap 出错
	MailServerError        = 1070
	InvalidAccessKeyId     = 1080 // AccessKeyId错误
	InvalidAccessKeySecret = 1081
	ForbiddenAccessKey     = 1082

	// 登录认证
	InvalidLogin       = 1110
	InvalidPassword    = 1111
	InvalidToken       = 1100 // 无效 token
	InvalidTokenScope  = 1101 // 无效 token scope
	InvalidTokenPrefix = 1102 // 无效 token prefix
	InvalidCaptcha     = 1103
	InvalidOperation   = 1104
	TokenExpired       = 1105
	InvalidOrgId       = 1106 // 无效的 orgId

	InvalidColumn = 1201
	DataToLong    = 1301
	NameToLong    = 1302
	RemarkToLong  = 1303

	// prometheus
	PromError      = 1400
	PromQueryError = 1410

	CloudAccountNoExists = 1504

	//  第三方接口错误
	CloudServerErr = 1501
	CMDBServerErr  = 1511

	// 用户
	UserAlreadyExists          = 2010
	UserNotExists              = 2020
	UserEmailDuplicate         = 2030
	UserEmailDuplicateInactive = 2031
	UserInvalidStatus          = 2040
	UserInactive               = 2041
	UserDisabled               = 2043
	InvalidPasswordFormat      = 2044 // 密码格式错误
	UserActivated              = 2045

	InvalidRoleName   = 3010
	RoleNameDuplicate = 3011

	// Customer
	InvalidCustomerKey = 3030
	CustomerNotExist   = 3031

	// process
	ProcessNotExists = 5010

	// datasource
	DsCheckError    = 6010
	DsNoCollInGroup = 6011

	// permission
	PermissionDeny  = 7010
	EmptyPermKey    = 7011
	InvalidPermName = 7012
	InvalidPermKey  = 7013
	EmptyPermApi    = 7014
	ValidateError   = 7015

	// license
	InvalidLicense        = 8010
	ExpiredLicense        = 8011
	LicenseResourceLimit  = 8020
	LicenseModuleDisabled = 8021

	// organization
	OrganizationAlreadyExists  = 9010
	OrganizationNotExists      = 9011
	OrganizationDisabled       = 9012
	OrganizationAliasDuplicate = 9013
	OrganizationInvalidStatus  = 9014
	InvalidOrganizationId      = 9015

	// consul
	ConsulConnError = 10010
)

var errorMsgs = map[int]map[string]string{
	InternalError: {
		"zh-cn": "未知错误",
	},
	ObjectAlreadyExists: {
		"zh-cn": "对象已存在",
	},
	ObjectNotExists: {
		"zh-cn": "对象不存在",
	},
	ObjectNotExistsOrNoPerm: {
		"zh-cn": "对象不存在或者无权限",
	},
	ObjectDisabled: {
		"zh-cn": "对象已禁用",
	},
	JSONParseError: {
		"zh-cn": "JSON 数据解析出错",
	},
	NotImplement: {
		"zh-cn": "暂未实现",
	},
	DBError: {
		"zh-cn": "数据库错误",
	},
	DBAttrValidateErr: {
		"zh-cn": "字段验证错误",
	},
	QuotaError: {
		"zh-cn": "超过额度限制",
	},
	BadParam: {
		"zh-cn": "无效参数",
	},
	BadRequest: {
		"zh-cn": "无效请求",
	},
	DataToLong: {
		"zh-cn": "数据过长",
	},
	NameToLong: {
		"zh-cn": "名称过长",
	},
	RemarkToLong: {
		"zh-cn": "备注过长",
	},
	IOError: {
		"zh-cn": "io 错误",
	},
	LdapError: {
		"zh-cn": "LDAP 错误",
	},
	MailServerError: {
		"zh-cn": "邮件服务错误",
	},
	InvalidAccessKeyId: {
		"zh-cn": "AccessKeyId错误",
	},
	InvalidAccessKeySecret: {
		"zh-cn": "AccessKeySecret错误",
	},
	ForbiddenAccessKey: {
		"zh-cn": "AccessKey权限不足",
	},
	InvalidToken: {
		"zh-cn": "凭证无效",
	},
	InvalidTokenScope: {
		"zh-cn": "凭证 scope 不匹配",
	},
	InvalidTokenPrefix: {
		"zh-cn": "凭证类型错误",
	},
	InvalidOrgId: {
		"zh-cn": "无效的组织",
	},
	TokenExpired: {
		"zh-cn": "凭证已过期",
	},
	PromError: {
		"zh-cn": "Prometheus 错误",
	},
	PromQueryError: {
		"zh-cn": "Prometheus 查询错误",
	},
	CloudAccountNoExists: {
		"zh-cn": "云商账号不存在",
	},
	StateTryAgainLater: {
		"zh-cn": "当前状态无法执行该操作，请稍后重试",
	},
	ColValidateError: {
		"zh-cn": "字段校验错误",
	},
	InvalidLogin: {
		"zh-cn": "无效的邮箱或密码",
	},
	InvalidPassword: {
		"zh-cn": "密码错误",
	},
	InvalidColumn: {
		"zh-cn": "无效的字段名",
	},
	InvalidCaptcha: {
		"zh-cn": "验证码错误",
	},
	InvalidOperation: {
		"zh-cn": "无效操作",
	},
	CMDBServerErr: {
		"zh-cn": "cmdb 接口错误",
	},
	UserAlreadyExists: {
		"zh-cn": "用户已存在",
	},
	UserNotExists: {
		"zh-cn": "用户不存在",
	},
	UserEmailDuplicate: {
		"zh-cn": "邮箱已注册，请直接登录",
	},
	UserEmailDuplicateInactive: {
		"zh-cn": "邮箱已注册，请前往邮箱激活账号",
	},
	UserInvalidStatus: {
		"zh-cn": "无效的用户状态",
	},
	UserInactive: {
		"zh-cn": "用户未激活",
	},
	UserDisabled: {
		"zh-cn": "用户已禁用",
	},
	InvalidPasswordFormat: {
		"zh-cn": "密码格式错误",
	},
	UserActivated: {
		"zh-cn": "账号已激活",
	},
	InvalidRoleName: {
		"zh-cn": "无效角色名",
	},
	RoleNameDuplicate: {
		"zh-cn": "角色名重复",
	},
	InvalidCustomerKey: {
		"zh-cn": "无效的客户 key",
	},
	CustomerNotExist: {
		"zh-cn": "客户不存在",
	},
	ProcessNotExists: {
		"zh-cn": "进程不存在",
	},
	DsCheckError: {
		"zh-cn": "监控源连接失败，请检查地址及认证信息",
	},
	DsNoCollInGroup: {
		"zh-cn": "指定分组中没有采集器",
	},
	PermissionDeny: {
		"zh-cn": "无权限",
	},
	EmptyPermKey: {
		"zh-cn": "权限 key 不能为空",
	},
	InvalidPermName: {
		"zh-cn": "无效的权限点名称",
	},
	InvalidPermKey: {
		"zh-cn": "无效的权限点 key",
	},
	EmptyPermApi: {
		"zh-cn": "权限点的 API 不能为空",
	},
	ValidateError: {
		"zh-cn": "验证失败",
	},
	InvalidLicense: {
		"zh-cn": "无效 License",
	},
	ExpiredLicense: {
		"zh-cn": "License 已过期",
	},
	LicenseResourceLimit: {
		"zh-cn": "资源数量超过 License 限制",
	},
	LicenseModuleDisabled: {
		"zh-cn": "功能模块未启用",
	},
	OrganizationAlreadyExists: {
		"zh-cn": "组织已存在",
	},
	OrganizationNotExists: {
		"zh-cn": "组织不存在",
	},
	OrganizationDisabled: {
		"zh-cn": "组织被禁用",
	},
	OrganizationAliasDuplicate: {
		"zh-cn": "组织别名重复",
	},
	OrganizationInvalidStatus: {
		"zh-cn": "无效的组织状态",
	},
	InvalidOrganizationId: {
		"zh-cn": "无效的组织ID",
	},
}
