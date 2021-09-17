// Copyright 2021 CloudJ Company Limited. All rights reserved.

package e

const (
	/*
		错误码定义规范:
		1. 错误码由 5 位组成: abcde
		2. a 代表一级分类，如通用、db、外部服务、业务逻辑、数据校验、权限等
		3. bc 代表二级分类, 一般对应具体功能模块，如用户、VCS、模板、环境等
		4. de 代表具体错误类型，如不存在、连接失败、验证失败、加解密错误等
	*/

	// 通用错误 1

	InternalError           = 10000
	ObjectAlreadyExists     = 10010
	ObjectNotExists         = 10011
	ObjectNotExistsOrNoPerm = 10013 // 对象不存在或者无权限
	ObjectDisabled          = 10014
	NotImplement            = 10020
	IOError                 = 10030 // 文件 io 出错

	//// 解析错误 101

	JSONParseError = 10100
	HCLParseError  = 10101

	//// db 错误 102

	DBError           = 10200 // db 操作出错
	DBAttrValidateErr = 10201
	ColValidateError  = 10202
	NameDuplicate     = 10203
	InvalidColumn     = 10210
	DataToLong        = 10211
	NameToLong        = 10212
	RemarkToLong      = 10213
	TagToLong         = 10214

	//// 校验错误 103

	BadParam               = 10340 // 参数错误(参数值不对)
	BadRequest             = 10341 // 请求错误(请求缺少必要参数)
	InvalidAccessKeyId     = 10380 // AccessKeyId错误
	InvalidAccessKeySecret = 10381
	ForbiddenAccessKey     = 10382

	//// 第三方服务错误 104

	LdapError       = 10410 // ldap 出错
	MailServerError = 10420
	ConsulConnError = 10430
	VcsError        = 10440

	// 权限认证 2
	//// 认证 200

	InvalidPassword   = 20010
	InvalidToken      = 20000 // 无效 token
	InvalidTokenScope = 20001 // 无效 token scope
	TokenExpired      = 20005
	InvalidOrgId      = 20006 // 无效的 orgId
	InvalidProjectId  = 20007 // 无效的 projectId

	//// 权限 201

	PermissionDeny   = 20110
	ValidateError    = 20111
	InvalidOperation = 20112

	// 功能模块 3
	//// 用户 301

	UserAlreadyExists          = 30110
	UserNotExists              = 30120
	UserEmailDuplicate         = 30130
	UserEmailDuplicateInactive = 30131
	UserInvalidStatus          = 30140
	UserInactive               = 30141
	UserDisabled               = 30143
	InvalidPasswordFormat      = 30144 // 密码格式错误
	UserActivated              = 30145
	InvalidRoleName            = 30150
	RoleNameDuplicate          = 30151

	//// 组织 303

	OrganizationAlreadyExists = 30310
	OrganizationNotExists     = 30311
	OrganizationDisabled      = 30312
	OrganizationInvalidStatus = 30314
	InvalidOrganizationId     = 30315

	//// project 304

	ProjectAlreadyExists      = 30410
	ProjectNotExists          = 30411
	ProjectAliasDuplicate     = 30412
	ProjectUserAlreadyExists  = 30420
	ProjectUserAliasDuplicate = 30421

	//// variable 305

	VariableAlreadyExists  = 30510
	VariableAliasDuplicate = 30511

	//// token 306

	TokenAlreadyExists  = 30610
	TokenNotExists      = 30611
	TokenAliasDuplicate = 30613

	//// template 307

	TemplateAlreadyExists   = 30710
	TemplateNotExists       = 30711
	TemplateDisabled        = 30712
	TemplateActiveEnvExists = 30730

	//// environment 308

	EnvAlreadyExists       = 30810
	EnvNotExists           = 30811
	EnvAliasDuplicate      = 30812
	EnvArchived            = 30813
	EnvCannotArchiveActive = 30814
	EnvDeploying           = 30815

	//// task 309

	TaskAlreadyExists     = 30910
	TaskNotExists         = 30911
	TaskApproveNotPending = 30913
	TaskStepNotExists     = 30914
	TaskNotHaveStep       = 30916

	//// ssh key 310

	KeyAlreadyExists  = 31010
	KeyNotExist       = 31011
	KeyAliasDuplicate = 31012
	KeyDecryptFail    = 31013

	//// vcs 311

	VcsNotExists   = 31110
	VcsDeleteError = 31120

	//// policy 312

	PolicyAlreadyExist           = 31210
	PolicyNotExist               = 31211
	PolicyGroupAlreadyExist      = 31221
	PolicyGroupNotExist          = 31222
	PolicyBelongedToAnotherGroup = 31223
	PolicyResultAlreadyExist     = 31230
	PolicyResultNotExist         = 31231
	PolicyRegoMissingComment     = 31410
	PolicyErrorParseTemplate     = 31510
	PolicySuppressNotExist       = 31610
	PolicySuppressAlreadyExist   = 31611
	PolicyRelNotExist            = 31710
	PolicyRelAlreadyExist        = 31711
	/// terraform 313

	InvalidTfVersion = 31300
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
	BadParam: {
		"zh-cn": "无效参数",
	},
	BadRequest: {
		"zh-cn": "无效请求",
	},
	DataToLong: {
		"zh-cn": "内容过长",
	},
	NameToLong: {
		"zh-cn": "名称过长",
	},
	RemarkToLong: {
		"zh-cn": "备注过长",
	},
	TagToLong: {
		"zh-cn": "标签过长",
	},
	IOError: {
		"zh-cn": "io 错误",
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
	InvalidOrgId: {
		"zh-cn": "无效的组织",
	},
	TokenExpired: {
		"zh-cn": "凭证已过期",
	},
	ColValidateError: {
		"zh-cn": "字段校验错误",
	},
	InvalidPassword: {
		"zh-cn": "无效的邮箱或密码",
	},
	InvalidColumn: {
		"zh-cn": "无效的字段名",
	},
	InvalidOperation: {
		"zh-cn": "无效操作",
	},
	UserAlreadyExists: {
		"zh-cn": "用户已存在",
	},
	UserNotExists: {
		"zh-cn": "用户不存在",
	},
	UserEmailDuplicate: {
		"zh-cn": "用户邮箱已存在",
	},
	UserEmailDuplicateInactive: {
		"zh-cn": "无效的用户邮箱",
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
	PermissionDeny: {
		"zh-cn": "无权限",
	},
	ValidateError: {
		"zh-cn": "验证失败",
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
	OrganizationInvalidStatus: {
		"zh-cn": "无效的组织状态",
	},
	InvalidOrganizationId: {
		"zh-cn": "无效的组织ID",
	},
	NameDuplicate: {
		"zh-cn": "名称重复",
	},
	TaskStepNotExists: {
		"zh-cn": "步骤不存在",
	},
	InvalidProjectId: {
		"zh-cn": "无效的项目id",
	},
	TaskNotHaveStep: {
		"zh-cn": "任务无步骤",
	},
	TemplateAlreadyExists: {
		"zh-cn": "模板名称重复",
	},
	HCLParseError: {
		"zh-cn": "模板语法解析错误",
	},

	VariableAliasDuplicate: {
		"zh-cn": "变量别名重复",
	},

	ProjectUserAlreadyExists: {
		"zh-cn": "项目用户已经存在",
	},

	ProjectUserAliasDuplicate: {
		"zh-cn": "项目别名重复",
	},

	TokenAlreadyExists: {
		"zh-cn": "Token已经存在",
	},
	TokenNotExists: {
		"zh-cn": "Token不存在",
	},
	TokenAliasDuplicate: {
		"zh-cn": "Token别名重复",
	},

	TemplateNotExists: {
		"zh-cn": "模板不存在",
	},
	TemplateDisabled: {
		"zh-cn": "模板不可用",
	},
	TemplateActiveEnvExists: {
		"zh-cn": "模板存在活跃环境",
	},
	ConsulConnError: {
		"zh-cn": "consul链接失败",
	},
	EnvAlreadyExists: {
		"zh-cn": "环境已经存在",
	},
	EnvNotExists: {
		"zh-cn": "环境不存在",
	},
	EnvAliasDuplicate: {
		"zh-cn": "环境别名重复",
	},
	EnvArchived: {
		"zh-cn": "环境已归档，不允许操作",
	},
	EnvDeploying: {
		"zh-cn": "环境正在部署中，请不要重复发起",
	},
	TaskAlreadyExists: {
		"zh-cn": "任务已经存在",
	},
	TaskNotExists: {
		"zh-cn": "任务不存在",
	},
	VcsError: {
		"zh-cn": "vcs仓库错误",
	},
	VcsNotExists: {
		"zh-cn": "vcs仓库不存在",
	},
	VcsDeleteError: {
		"zh-cn": "vcs存在相关依赖云模版，无法删除",
	},
	TaskApproveNotPending: {
		"zh-cn": "作业状态非待审批，不允许操作",
	},
	KeyAlreadyExists: {
		"zh-cn": "管理秘钥已存在",
	},
	KeyNotExist: {
		"zh-cn": "管理秘钥不存在",
	},
	KeyAliasDuplicate: {
		"zh-cn": "管理秘钥名称重复",
	},
	KeyDecryptFail: {
		"zh-cn": "管理秘钥解析失败",
	},
	EnvCannotArchiveActive: {
		"zh-cn": "环境当前状态活跃, 无法归档",
	},
	InvalidTfVersion: {
		"zh-cn": "自动选择版本失败",
	},

	PolicyAlreadyExist: {
		"zh-cn": "策略已存在",
	},

	PolicyNotExist: {
		"zh-cn": "策略不存在",
	},

	PolicyGroupAlreadyExist: {
		"zh-cn": "策略组已存在",
	},

	PolicyGroupNotExist: {
		"zh-cn": "策略组不存在",
	},

	PolicyBelongedToAnotherGroup: {
		"zh-cn": "策略属于其他策略组",
	},

	PolicyResultAlreadyExist: {
		"zh-cn": "结果已存在",
	},

	PolicyResultNotExist: {
		"zh-cn": "结果不存在",
	},

	PolicyErrorParseTemplate: {
		"zh-cn": "模板解析错误",
	},

	PolicyRegoMissingComment: {
		"zh-cn": "Rego脚本头缺失",
	},

	PolicySuppressNotExist: {
		"zh-cn": "屏蔽记录不存在",
	},

	PolicySuppressAlreadyExist: {
		"zh-cn": "屏蔽记录已存在",
	},

	PolicyRelNotExist: {
		"zh-cn": "策略关联关系不存在",
	},

	PolicyRelAlreadyExist: {
		"zh-cn": "策略关联关系已存在",
	},
}
