------
## v0.6.0-pre9 (待发版)
#### Fixed
- 后端中文字符长度判断错误(jira#cloudiac-30)
- 环境合规状态没有在合规步骤执行完成后及时统计(jira#cloudiac-31)


------
## v0.6.0-pre8 20210922
#### Refactor
- 调整创建部署任务时判断是否创建策略扫描步骤的逻辑，保证步骤索引连续

------
## v0.6.0-pre7 20210918
#### Features
- 环境绑定策略组或启用检测时显示云模板绑定的策略组列表

#### Changes
- 合规概览，未解决错误策略只统计所有环境或云模板最后一次执行的数据
- 合规策略详情，策略错误列表返回环境或云模板最后一次失败列表，最新的排第一个
- 合规策略详情，图表中没有扫描记录的日期也应该显示该天的时间
- 合规策略详情，通过率趋势改为显示百分比

#### Fixed
- 环境部署过程中允许删除关联云模板(jira#cloudiac-1)
- 存在活跃环境的云模板在列表中活跃资源数显示为0(jira#cloudiac-9)
- 云模板取消所有策略组绑定 500 错误
- 最近 n 天时间计算不错误

------
## v0.6.0-pre5 20210917
#### Changes
- 环境自动继承云模板绑定的策略组，在执行检测时计算环境生效的所有策略组
- 任务执行失败时策略不再被标记为失败
- 云模板按组织创建时间+模板创建时间排序
- 环境按组织创建时间+项目创建时间+环境创建时间排序

#### Fxied
- 环境检测弹窗中显示上一次检测结果且不会自动刷新(jira#cloudiac-22)
- 任务评论超长提示不友好(jira#cloudiac-7)
- 策略标签长度限制为 16 个字符(jira#cloudiac-25)
- 未开启检测的云模板显示状态为通过(jira#cloudiac-20)

-----
## v0.6.0-pre4 20210915
#### Fxied
- 修复无法销毁环境的问题(jira#CLOUDIAC-24)
- 修复环境部署时执行了合规检查，但合规状态无数据的问题
- 修复创建新环境时未同步云模板合规设置的问题

-----
## v0.6.0-pre2 20210914
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
- 环境增加 lastResTaskId 字段，记录最后一次可能进行了资源改动的任务 id，
避免任务被驳回时环境的资源数量统计为 0 的问题。

#### Changes
- 使用 TF_VAR_xxx 格式的环境变量进行 terraform 变量的传入，避免传入未声明的变量时出现警告信息。

#### 升级步骤
1. 备份数据库
2. 更新并重启后执行以下 SQL
```
UPDATE iac_env SET last_res_task_id=last_task_id WHERE last_res_task_id IS NULL;
```

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

