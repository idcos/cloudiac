// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

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
	TooManyRetries          = 10040
	EncryptError            = 10050
	DecryptError            = 10051

	//// 解析错误 101
	JSONParseError = 10100
	HCLParseError  = 10101
	URLParseError  = 10102

	//// db 错误 102
	DBError           = 10200 // db 操作出错
	DBAttrValidateErr = 10201
	ColValidateError  = 10202
	NameDuplicate     = 10203
	InvalidColumn     = 10210
	DataTooLong       = 10211
	NameTooLong       = 10212
	RemarkTooLong     = 10213
	TagTooLong        = 10214
	TagTooMuch        = 10215

	//// 校验错误 103
	BadOrgId               = 10310
	BadProjectId           = 10311
	BadTemplateId          = 10312
	BadEnvId               = 10314
	BadParam               = 10340 // 参数错误(参数值不对)
	BadRequest             = 10341 // 请求错误(请求缺少必要参数)
	InvalidPipeline        = 10350
	InvalidPipelineVersion = 10351
	InvalidExportVersion   = 10361
	InvalidAccessKeyId     = 10380 // AccessKeyId错误
	InvalidAccessKeySecret = 10381
	ForbiddenAccessKey     = 10382
	TemplateNameRepeat     = 10383
	TemplateWorkdirError   = 10384

	//// 第三方服务错误 104
	LdapError       = 10410 // ldap 出错
	MailServerError = 10420
	ConsulConnError = 10430

	// vcs调用相关错误
	VcsError          = 10440
	VcsAddressError   = 10441
	VcsInvalidToken   = 10442
	VcsConnectError   = 10445
	VcsConnectTimeOut = 10446

	//// 导入导出错误 105
	ImportError       = 10510
	ImportIdDuplicate = 10520 //  id 重复
	ImportUpdateOrgId = 10530

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
	PermDenyApproval = 20113

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
	VariableScopeConflict  = 30512
	InvalidVarName         = 30513
	EmptyVarName           = 30514
	EmptyVarValue          = 30515

	//// token 306
	TokenAlreadyExists  = 30610
	TokenNotExists      = 30611
	TokenAliasDuplicate = 30613

	//// template 307
	TemplateAlreadyExists   = 30710
	TemplateNotExists       = 30711
	TemplateDisabled        = 30712
	TemplateActiveEnvExists = 30730
	TemplateKeyIdNotSet     = 30731

	//// environment 308
	EnvAlreadyExists       = 30810
	EnvNotExists           = 30811
	EnvAliasDuplicate      = 30812
	EnvArchived            = 30813
	EnvCannotArchiveActive = 30814
	EnvDeploying           = 30815
	EnvCheckAutoApproval   = 30816
	EnvTagNumLimited       = 30821
	EnvTagLengthLimited    = 30822

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

	//// 317
	RegistryServiceErr = 31710

	//// policy 312
	PolicyAlreadyExist           = 31210
	PolicyNotExist               = 31211
	PolicyGroupAlreadyExist      = 31221
	PolicyGroupNotExist          = 31222
	PolicyBelongedToAnotherGroup = 31223
	PolicyResultAlreadyExist     = 31230
	PolicyResultNotExist         = 31231
	PolicyRegoMissingComment     = 31340
	PolicyErrorParseTemplate     = 31250
	PolicySuppressNotExist       = 31260
	PolicySuppressAlreadyExist   = 31261
	PolicyRelNotExist            = 31270
	PolicyRelAlreadyExist        = 31271
	PolicyScanNotEnabled         = 31280
	PolicyMetaInvalid            = 31281
	PolicyRegoInvalid            = 31282
	PolicyGroupDirError          = 31283

	/// terraform 313
	InvalidTfVersion = 31300

	// VariableGroup 314

	VariableGroupAlreadyExist   = 31410
	VariableGroupNotExist       = 31411
	VariableGroupAliasDuplicate = 31412

	//cron 315
	CronExpressError = 31500
	CronTaskFailed   = 31501

	// system config 316
	SystemConfigNotExist = 31610
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
	URLParseError: {
		"zh-cn": "URL解析出错",
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
	BadOrgId: {
		"zh-cn": "组织 ID 错误",
	},
	BadProjectId: {
		"zh-cn": "项目 ID 错误",
	},
	BadTemplateId: {
		"zh-cn": "模板 ID 错误",
	},
	BadEnvId: {
		"zh-cn": "环境 ID 错误",
	},
	BadParam: {
		"zh-cn": "无效参数",
	},
	TemplateNameRepeat: {
		"zh-cn": "云模版名称重复",
	},
	TemplateWorkdirError: {
		"zh-cn": "工作目录校验失败",
	},
	BadRequest: {
		"zh-cn": "无效请求",
	},
	InvalidPipeline: {
		"zh-cn": "pipeline 格式错误",
	},
	InvalidPipelineVersion: {
		"zh-cn": "不支持的 pipeline 版本",
	},
	InvalidExportVersion: {
		"zh-cn": "不支持的导出数据版本",
	},
	DataTooLong: {
		"zh-cn": "内容过长",
	},
	NameTooLong: {
		"zh-cn": "名称过长",
	},
	RemarkTooLong: {
		"zh-cn": "备注过长",
	},
	TagTooLong: {
		"zh-cn": "标签过长",
	},
	TagTooMuch: {
		"zh-cn": "标签过多",
	},
	IOError: {
		"zh-cn": "io 错误",
	},
	TooManyRetries: {
		"zh-cn": "达到最大重试次数",
	},
	EncryptError: {
		"zh-cn": "数据加密错误",
	},
	DecryptError: {
		"zh-cn": "数据解密错误",
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
	PermDenyApproval: {
		"zh-cn": "无审批权限",
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

	VariableAlreadyExists: {
		"zh-cn": "变量已存在",
	},
	VariableAliasDuplicate: {
		"zh-cn": "变量别名重复",
	},
	VariableScopeConflict: {
		"zh-cn": "变量作用域冲突",
	},
	InvalidVarName: {
		"zh-cn": "无效变量名",
	},
	EmptyVarName: {
		"zh-cn": "变量名不可为空",
	},
	EmptyVarValue: {
		"zh-cn": "变量值不可为空",
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
	EnvCheckAutoApproval: {
		"zh-cn": "配置自动纠漂移、推送到分支时重新部署时，必须配置自动审批",
	},
	TaskAlreadyExists: {
		"zh-cn": "任务已经存在",
	},
	EnvTagNumLimited: {
		"zh-cn": "环境 tag 数量超过限制",
	},
	EnvTagLengthLimited: {
		"zh-cn": "环境 tag 长度超过限制",
	},
	TaskNotExists: {
		"zh-cn": "任务不存在",
	},
	VcsError: {
		"zh-cn": "vcs仓库错误",
	},
	VcsAddressError: {
		"zh-cn": "vcs地址错误",
	},
	VcsInvalidToken: {
		"zh-cn": "vcs token无效",
	},
	VcsConnectError: {
		"zh-cn": "vcs服务连接失败",
	},
	VcsConnectTimeOut: {
		"zh-cn": "vcs服务连接超时",
	},
	VcsNotExists: {
		"zh-cn": "vcs仓库不存在",
	},
	VcsDeleteError: {
		"zh-cn": "vcs存在相关依赖云模版，无法删除",
	},
	ImportError: {
		"zh-cn": "导入出错",
	},
	ImportIdDuplicate: {
		"zh-cn": "id 重复",
	},
	ImportUpdateOrgId: {
		"zh-cn": "同 id 的数据己属于另一组织，无法使用“覆盖”方案(不允许更改组织 id)",
	},
	TaskApproveNotPending: {
		"zh-cn": "作业状态非待审批，不允许操作",
	},
	KeyAlreadyExists: {
		"zh-cn": "管理密钥已存在",
	},
	KeyNotExist: {
		"zh-cn": "管理密钥不存在",
	},
	KeyAliasDuplicate: {
		"zh-cn": "管理密钥名称重复",
	},
	KeyDecryptFail: {
		"zh-cn": "管理密钥解析失败",
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

	PolicyScanNotEnabled: {
		"zh-cn": "扫描未启用",
	},
	CronExpressError: {
		"zh-cn": "cron定时任务表达式错误",
	},
	CronTaskFailed: {
		"zh-cn": "cron定时任务执行失败",
	},
	PolicyMetaInvalid: {
		"zh-cn": "策略元数据解析无效",
	},
	PolicyRegoInvalid: {
		"zh-cn": "rego 脚本解析无效",
	},
	SystemConfigNotExist: {
		"zh-cn": "当前配置不存在",
	},
	TemplateKeyIdNotSet: {
		"zh-cn": "SSH 密钥未配置",
	},
	PolicyGroupDirError: {
		"zh-cn": "仓库在当前目录找不到策略文件",
	},
}
