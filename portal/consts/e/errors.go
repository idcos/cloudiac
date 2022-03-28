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

	//// runner 报错 106
	RunnerError = 10610

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
	EnvAlreadyExists        = 30810
	EnvNotExists            = 30811
	EnvAliasDuplicate       = 30812
	EnvArchived             = 30813
	EnvCannotArchiveActive  = 30814
	EnvDeploying            = 30815
	EnvCheckAutoApproval    = 30816
	EnvLockFailedTaskActive = 30817
	EnvTagNumLimited        = 30821
	EnvTagLengthLimited     = 30822

	//// task 309
	TaskAlreadyExists     = 30910
	TaskNotExists         = 30911
	TaskApproveNotPending = 30913
	TaskStepNotExists     = 30914
	TaskNotHaveStep       = 30916
	TaskAborting          = 30917
	TaskCannotAbort       = 30918

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

	// Ldap 317
	LdapConnectFailed = 31710
	LdapUpdateFailed  = 31720
)
