## v1.3.7.1 20230925
### Features:
- 新增privileged配置

------
## v1.3.7 20230426
### Features:
- 新增iac部署日志的错误信息接口

------
## v1.3.6 20230323
### Fixed
- 修复资源列表无法区分datasource和resource的问题
- 修复ansible软连接无法读取的问题

#### 升级步骤
**升级前注意备份数据**

**数据升级**
执行 `./iac-tool updateDB resource -m` 进行数据升级。
如果是容器化部署执行: `docker-compose exec iac-portal ./iac-tool updateDB resource -m`

------
## v1.3.5 20230208
### Features
- 新增swaggerEnable参数支持开关swagger文档
### Fixed
- 修复事件通知时分支显示 stack 分支名称
### Changed
- 优化 checkOrgId 时数据库错误的报错

------
## v1.3.4 20221224
### Fixed
- 修复删除历史漂移检测任务耗时过长的问题
- 修复重新部署时无法真正删除 ssh 密钥的问题
- 修复删除资源账号引用关系时其所有引用关系都被清空的问题
- 修复 gitlab 仓库的分支和标签只能获取 100 个的问题
- 修复环境定时销毁或部署失败后无法再创建新的定时任务的问题
- 修复步骤默认超时时间单位错误的问题

------
## v1.3.3 20221222
### Fixed:
- 修复定时销毁/部署任务失败后，重新设置无法进行定时销毁/部署任务后失效的问题
- 修复步骤超时时间转换出错问题

------
## v1.3.2 20221215
### Features:
- 漂移检测后发送 kafka 消息
- kafka回调增加 output 信息
- 修改 plan result 不返回值的问题
- 默认使用的 terraform 版本更新为 1.2.4

### Changed
- 修改任务下发runner选择策略为轮循
- 优化 gorm 日志中输出的文件名和路径
- 更新 base-image 到 0.2.1

### Fixed
- 修复资源账号分页无效（只显示单页）的问题
- 修复环境存活时间设置为周期后无法重新设置为其他模式的问题
- 修复环境存活时间设置为周期时重新部署后周期销毁表达式被清空的问题
- 创建部署任务时可能出现死锁的问题


------
## v1.2.2 20221014
### Fixed
- 修复 terraform.py 脚本有中文注释导致运行出错的问题

------
## v1.2.1 20220914
### Fixed
- 修复 _cloudiac.tfvars.json 文件生成问题
- _cloudiac.tfvars.json 无法处理复杂变量(json)的问题
- 修复传入 TF_LOG 环境变量导致 terraform.py 脚本出错的问题(again)
- 修复环境部署日志问题

------
## v1.2.0 20220906
### Features
- ct-worker 镜像添加 cloudiac-playbook 命令

### Fixed
- 修复漂移检测数组下标越界问题

### Enhancements
- 改用 tfvars.json 文件传入 terraform 变量
- 环境部署时未传workdir则使用stack的workdir


------
## v1.1.0 20220819
#### Features
- 新增环境存活时间周期维度，通过crontab表达式设置环境定时部署，定时销毁

#### Fixes
- 修复修改环境配置时，环境标签被删除的问题

------
## v1.0.0 20220714
#### Features
- 新增pipeline v0.5
- 新增平台概览
- 新增用户操作日志
  - 用户登录
  - 环境部署、销毁、重新部署
  - 组织、项目、Stack创建
- 新增用户找回密码功能
  - 找回密码邮件通知
  - 修改密码
- 环境搜索新增多状态查询，按环境修改时间查询
- 内置 terraform 增加 v1.1.9/v1.2.4 版本
- 新增Stack创建来源，从Exchange创建

#### Enhancements
- 优化敏感变量展示
  1. 环境页面中 Output tab下的敏感变量隐藏显示
  2. 环境页面中 资源详情的 json 描述中，敏感变量隐藏显示
  
#### Fixes
- 修复邮件通知重复发送问题
- 修复任务详情变量显示问题
- 修复屏蔽策略失败的问题
- 修复按策略屏蔽时数据查询异常问题
- 修复云模板列表中的repo地址展示异常问题
- 修复vcs分支/标签 默认只能返回20问题
- 修复合规检测解析行数与文件失败问题

------
## v0.12.1 20220625
#### Fixes
- 修复项目审批员的访问权限问题

------
## v0.12.0 20220624
#### Features
- 新增用户注册功能
- kafka回调消息增加任务id和合规状态字段
- resource表中增加依赖关系字段
- 资源查询接口支持展示依赖数据
- 获取ldap ou信息时增加加上组织过滤

#### Changes
- 『云模板』统一更名为『Stack』

#### Enhancements
- 对任务入参做处理，避免 shell 注入
    - 存在可能导入注入的入参时直接报错
    - 针对 targets 参数会做 shellescape 处理
- 密钥管理支持RSA密钥
- 优化注册密码强度
    - 字母、数字、符号至少包含两种，长度6-30
- 优化任务详情及output接口鉴权逻辑
- 优化vcs相关查询接口权限
- 优化环境名称重复时的提示信息
- 优化token输出，不在屏蔽已过期的token
- 优化Stack相关查询权限

#### Fixes
- 修复plan后直接部署、销毁时 workdir 问题
- 修复使用Stack导入功能，传入 json结构 时 panic 问题
- 修复趋势费用缺少2月份数据的问题
- 组织概览左下角图的统计维度调整
- 修复环境标签只剩一个的情况下无法删除的问题
- 修复敏感变量加密问题
- 修复合规检测错误,列表显示问题
- 修复 Stack 的仓库地址展示错误的问题
- 修复 gitlab vcs token 验证失败时的报错
- 修复 Stack 敏感变量输出了 value 的问题
- 修复 portal 的 rego runtime 会暴露敏感环境变量的问题
- 修复环境变量和 terraform 变量有同名时无法删除的问题
- task/scantask/variable/varGroup/vcs 相关接口返回做脱敏处理
- 修复解析时间类型字符串异常问题
- 系统参数设置字符长度限制
- 重新部署/编辑环境存活时间更新失败
- 环境部署时，设置工作目录为空不生效的问题
- 修复继承的变量可以被删除的问题

#### 升级步骤
**升级前注意备份数据**

**数据升级**
执行 `./iac-tool updateDB resource -d` 进行数据升级。
如果是容器化部署执行: `docker-compose exec iac-portal ./iac-tool updateDB resource -d`

------
## v0.11.0 20220530
#### Features
- 组织和项目支持批量邀请用户
- 环境的部署通道改为使用 runner tag 进行匹配
- 新增，支持 ldap 登陆
- 新增，环境锁定功能
- 新增，环境归档后名称添加后缀
- 新增，环境支持设置工作目录
- 新增，环境和系统设置中增加部署任务的步骤超时时间设置(默认 60 分钟)
- 组织资源查询支持搜索资源属性关键字及环境和 provider 过滤
- 新增，组织和项目概览页统计数据
- 新增，aliyun 资源费用采集
- 新增环境概览页面，展示环境详情和统计数据
- 新增，资源账号增加项目关联及 provider 和费用统计选项
- 新增，审批时展示资源变更数据
- 项目管理员新增组织用户邀请权限
- 新增价格预估功能，在审批部署任务时展示资源变更的预估费用情况
- 增加 registry network mirror 支持，配置了 registry 服务地址后会自动启用该地址作为 network mirror server
- 接入自研 cloudcost 询价服务，目前支持的产品 aliyun ecs/disk/nat/slb/eip/rds/redis/mongodb
- 支持 ldap ou 或用户预授权
- 支持 consul 开启 acl 认证

#### Enhancements
- vcs 创建或编辑保存时对 token 有效性做校验
- 漂移检测任务运行时会再次获取环境最新一次执行部署时使用的配置信息
- 优化环境可部署状态检查接口，现在只检查环境关联的云模板 和 vcs 是否有效
- 任务类型为 apply 但未执行 terraformApply 步骤的漂移检测任务在部署历史中不展示
- 优化自动纠偏任务执行完 plan 后会判断是否有漂移，若无漂移则提前结束任务
- 优化 vcs 连接失败时的报错
- 组织层级的云模板列表接口返回关联环境数量
- 环境增加已销毁状态
- consul 服务创建锁并自动重新注册

#### Fixes
- 修复设置工作目录后 tfvars 和 playbook 文件路径保存错误的问题
- 修复 playbook 中输出中文内容会乱码的问题
- 修复开启 terraform debug 时执行 playbook 会报错的问题
- 修复工作目录不支持二层以上子目录的问题
- 修复导入云模板时因为没有选择关联项目导致报错的问题
- 修复云模板中的敏感变量导入后变为乱码的问题
- 修复查询资源账号时敏感变量值没有返回为空的问题
- 修复项目列表默认展示了已归档项目的问题
- 修复创建云模板或者测试策略组不绑定项目会报错的问题
- 修复环境锁定后 plan 完成可以发起部署的问题


#### 配置更新
若要开启 consul acl访问，需在 .env 中添加(不配置默认为 false)
```
CONSUL_ACL=true
CONSUL_ACL_TOKEN=""
```

ldap配置，需在 .env 中添加
```
LDAP_ADMIN_DN=""
LDAP_ADMIN_PASSWORD=""
LDAP_SERVER=""
LDAP_SEARCH_BASE=""
SEARCH_FILTER=""
EMAIL_ATTRIBUTE=""
ACCOUNT_ATTRIBUTE=""
```

询价服务配置，需在 .env 中添加
```
COST_SERVE=""
```


#### BREAKING CHANGES
- 旧环境配置的 runner id 替换为 tags
- 为旧的 iac_resource 数据自动设置 res_id 和 applied_at 值
- iac_task新增字段 applied
- 设置旧环境的工作目录为其使用的云模板的工作目录

#### 升级步骤
**升级前注意备份数据**

**数据升级**
执行 `iac-tool upgrade2v0.10` 进行数据升级。
如果是容器化部署执行: `docker-compose exec iac-portal ./iac-tool upgrade2v0.10`

**SQL 更新:**
```sql
-- iac_task新增字段 applied
update iac_task set applied = 1 where id in (select task_id from      iac_task_step where type = 'terraformApply' and status != 'pending') ;

-- 设置旧环境的工作目录为其使用的云模板的工作目录(只能升级后执行一次)
update iac_env join iac_template on iac_env.tpl_id = iac_template.id set iac_env.workdir = iac_template.workdir where iac_env.workdir = '';
```


------
## v0.9.4 20220414
#### Features
- 任务结束后的 kafka 回调消息中增加任务类型和环境状态
- 销毁资源、重新部署接口增加 source 字段，第三方服务调用时可通过该字段设置触发来源


------
## v0.9.3 20220323
#### Fixes
- 修复策略组查询时未正确处理 scope 导致创建环境报错的问题


------
## v0.9.2 20220318
#### Fixes
- 修复创建环境时 sampleVariable 变量会重复添加的问题


------
## v0.9.1 20220310
#### Features
- 环境支持设置及展示标签
- 环境创建、销毁、重新部署时都发送 kafka 消息，通知环境最新资源数据


------
## v0.9.0 20220307
#### Features
- 合规策略组改用代码库进行管理，支持通过分支或 tag 来管理版本
- 增强合规检测引擎，细化云模板及环境检测流程
- 执行界面增加云模板和环境的合规开关和合规策略组绑定功能
- 新增合规管理员角色
- 云模板新增 ssh 密钥配置
- 新增环境搜索功能，支持通过环境名称和云模板名称进行搜索
- 环境部署历史增加触发类型字段，记录部署任务的触发来源
- 环境部署历史页面增加合规状态展示
- 新增 iac registry 地址配置，环境变量 REGISTRY_ADDRESS
- 新增支持添加 iac registry 中发布的合规策略组
- 新增 HTTP_CLIENT_INSECURE 配置，默认为 false
- 增加 DOCKER_REGISTRY 环境变量配置，允许自定义 docker registry 地址

#### Enhancements
- 优化 docker clinet 连接调用，避免占用过多文件句柄
- 启用代码质量检查，修复代码质量问题(360+)
- 优化 vcs 服务报错，将 vcs 错误进一步细分为连接错误、认证错误等
- 项目云模板列表的“活跃环境”字段改为关联环境，点击数字可跳转到环境列表页面
 
#### Fixes
- 修复设置工作目录后无法选择工作目录下的 ansible plabyook 和 tfvars 文件的问题
- 修复组织管理员无权修改项目名称和描述的问题
- 修复合规策略中的 @name 注释未生效的问题 
- 修复任务 plan 失败后可能长时间不退出的问题
- 修复第三方系统调用接口创建的环境总是使用 master 分支的问题
- 修复部署环境资源账号继承问题
- 修复项目下添加资源账号校验异常问题
- 修复合规扫描过程中修改策略组绑定导致扫描结果异常问题
- 修复 gitee 私有仓库无法认证的问题
- 修复导出云模板时资源账号的敏感变量加解密处理错误的问题
- 修复任务驳回后状态显示为“失败”的问题
- 修复触发器触发归档环境部署问题
- 修复部分查询未正常处理软删除的问题
- 修复 runner 未处理非 32 位 secretKey 的问题
- 修复 command 步骤 cd 目录失效的问题
- 修复 local vcs 创建的云模板，仓库地址显示错误的问题

#### BREAKING CHANGES
- API 接口不再支持处理 GET 请求的 body
- 本次发版对合规策略的管理进行了重新设计，旧版本的合规策略数据不再支持，所有合规策略需要重新导入

#### 升级步骤
**升级前注意备份数据**

**SQL 更新:**
```sql
-- 合规数据清理
DROP TABLE `iac_policy`;
DROP TABLE `iac_policy_group`;
DROP TABLE `iac_policy_rel`;
DROP TABLE `iac_policy_result`;
DROP TABLE `iac_policy_suppress`;
DROP TABLE `iac_scan_task`;

-- 清空last_scan_task_id
UPDATE `iac_env` SET `last_scan_task_id` = '';
UPDATE `iac_template` SET `last_scan_task_id` = '';

-- 确让字段格式正确
ALTER TABLE `iac_task` CHANGE `message` `message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE `iac_user_org` CHANGE `role` `role` enum('admin','member','complianceManager');
```


------
## v0.8.1 20211214
#### Fixes
- 修复新组织中创建环境时接口报错的问题
- 修复有敏感变量时环境执行部署报解密错误的问题
- 修复执行任务容器异常退出会导致任务一直处于执行中状态且环境的资源一直累积的问题

#### Changes
- 编辑云模板时如果未传 repoAddr 则会清空 repoAddr，以支持修改旧模块的仓库


------
## v0.8.0 20211210
#### Features
- 新增环境漂移检测功能
- 新增环境资源图形化展示
- 新增云模板导出导入功能
- 新增创建云模板时名称和工作目录有效性检查
- 新增加 VCS 编辑功能，并对 vcs token 做加密处理
- 新增, 执行 MR/PR 触发的 plan 任务时将日志回写到 review comment
- 任务通知消息中增加任务类型说明
- 接口支持传参数 pageSize=0 表示不分页
- runner 启动 worker 前先进行 docker image pull

#### Fixes
- 修复步骤超时后不显示日志的问题
- 修复合规中心的云模板列表显示了己删除云模板的问题
- 修复合规检测任务 runner 未正常初始化策略文件导致合规检测总是成功的问题
- 修复查询 vcs 列表时无排序导致分页时可能出现重复数据
- 修复 vcs webhook 触发的任务创建人为空的问题
- 修复 github token 验证异常的问题
- 修复编辑云模板时仓库名称可能显示为 id 的问题
- 修复创建环境的 sampleVariable 入参处理逻辑错误导致创建重复变量的问题
- 修复合规检测任务的容器不会被中止的问题

#### Changes
- 修改步骤默认超时时间为 1800 秒
- 变量更新接口改为只支持传当前实例添加的变量
- 环境销毁时总是使用最后一次部署(apply)任务的 commit id
- 变量按名称排序
- 云模板选择仓库时仅列出与 token 用户相关的仓库(gitea 修改，其他 VCS 无此问题)
- 文档中默认使用的 mysql 版本修改为 8.0


------
## v0.7.1 20211117
#### Features
- 新增 runner 的 offline mode

#### Enhancements
- 调整步骤的默认超时时间为 1800 秒
- 步骤超时会记录错误原因，展示为 "timeout"

#### Fixes
- 修复 ct-worker 镜像的 provider 加载问题


#### 配置更新
若要开启 offline mode，需在 .env 中添加(不配置默认为 false) 
```
RUNNER_OFFLINE_MODE="true"
```


------
## v0.7.0 20211105
#### Features
- **新增自定义 pipeline 功能，并将任务执行过程分步展示**
- 新增组织内资源查询功能
- 新增资源账号管理功能
- 新增 kafka 任务执行结果回调通知

#### Enhancements
- 优化从组织和项目中移除用户功能
- 组织中编辑用户时允许修改姓名和手机号

#### Fixes
- 修复从组织中删除用户后用户在项目中依然存在的问题
- 修复设置环境自动触发 plan/apply 功能报错的问题
- 修复 local vcs 的文件搜索实现总是会递归查找文件的问题


------
## v0.6.1 20211027
#### Changes
- 更新 docker 镜像打包方案，先打包 base image，再基于 base image 构建最终镜像


------
## v0.6.0 20210928
#### Features
- **新增合规检查功能，平台管理员可进行合规管理**；
- 新增消息通知功能；
- 新增自动设置 vcs webhook 功能；
- 新增任务重试功能，可在环境设置中开启执行失败自动重试；
- 新增 tfvars 文件和 playbook 文件内容查看功能；
- 新增 terraform 版本选择功能，并支持自动匹配；
- 新增环境的资源详情展示，点击资源名称可查看资源详情；
- 新增选择型变量支持；
- 删除 VCS 时进行依赖检查，有模板依赖 VCS 时不允许删除；
- 任务无资源变更数据时“资源变更”字段不展示数值(避免展示为 0)；
- 任务增加审批驳回状态，审批驳回不再显示为“失败”。

#### Fixed
- 环境部署过程中允许删除关联云模板
- 存在活跃环境的云模板在列表中活跃资源数显示为0
- 任务评论超长提示不友好的问题

#### Changes
- 使用 TF_VAR_xxx 格式的环境变量进行 terraform 变量的传入，避免传入未声明的变量时出现警告信息。
- 环境增加 lastResTaskId 字段，记录最后一次可能进行了资源改动的任务 id，
避免任务被驳回时环境的资源数量统计为 0 的问题。

#### 升级步骤
1. 备份数据库
2. 更新并重启后执行以下 SQL
```
UPDATE iac_env SET last_res_task_id=last_task_id WHERE last_res_task_id IS NULL;
```

*可以跳过 v0.5.1 直接升级到该版本，但需要确保执行 v0.5.1 升级步骤中的 SQL*

------
## v0.5.1 20210806
#### Features
- 支持配置 JWT 和 AES 的密钥

#### Fixed
- 修复有敏感变量时执行部署报错的问题
- 修复无组织时返回 nil 导致前端报错的问题
- 修复部署日志添加评论报“任务己存在”的问题
- 修复 local VCS 分支中带 "/" 时无法正常处理的问题

#### Changes
- 升级 gorm2.0
- 修改 repos/cloud-iac 为 repos/cloudiac
- 模板的 tfvars 和 playbook 配置只在创建环境时使用，之后模板的修改不影响环境
- 调整任务队列和任务状态的轮询间隔为 1s
- 环境非活跃时设置 ttl 不同步设置 autoDestroyAt

#### 升级步骤
1. 升级完成后执行以下 sql，更新模板的 repo_id 和 repo_addr        
**备份数据**
```sql
update iac_template SET repo_id = replace(repo_id,'/cloud-iac/','/cloudiac/') where repo_id like '/cloud-iac/%';
update iac_template SET repo_addr = replace(repo_addr,'/repos/cloud-iac/','/repos/cloudiac/') where repo_addr like '%/repos/cloud-iac/%';
```

2. 删除 deleted_at 字段
gorm 统一使用了 deleted_at_t 字段进行软删除标识

**备份数据**
```sql
ALTER TABLE iac_env DROP COLUMN deleted_at;
ALTER TABLE iac_task DROP COLUMN deleted_at;
ALTER TABLE iac_user DROP COLUMN deleted_at;
ALTER TABLE iac_project DROP COLUMN deleted_at;
ALTER TABLE iac_template DROP COLUMN deleted_at;
```

3. 添加 SECRET_KEY 环境变量配置
`.env` 文件中添加以下内容(若己存在则不需要配置)
```
SECRET_KEY=xxxx	# 变量值请根据环境进行设置
```


------
## v0.5.0 20210728
全新 0.5.0 版本发布
