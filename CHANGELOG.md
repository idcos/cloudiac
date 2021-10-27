------
## v0.7.0-rc1 20211027
#### Features
- **新增自定义 pipeline 功能，并将任务执行过程分步展示**
- 新增组织内资源查询功能
- 新增资源账号管理功能

#### Enhancements
- 优化从组织和项目中移除用户功能

#### Fixes
- 修复从组织中删除用户后用户在项目中依然存在的问题
- 修复设置环境自动触发 plan/apply 功能报错的问题


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

