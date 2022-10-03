# Releases

------
## v1.2.1 20220914
**Fixes**

- 修复 _cloudiac.tfvars.json 文件生成问题
- _cloudiac.tfvars.json 无法处理复杂变量(json)的问题
- 修复传入 TF_LOG 环境变量导致 terraform.py 脚本出错的问题(again)
- 修复环境部署日志问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v1.2.1](https://github.com/idcos/cloudiac/releases/tag/v1.2.1)


------
## v1.2.0 20220906
**Enhancements**

- 改用 tfvars.json 文件传入 terraform 变量
- 环境部署时未传workdir则使用stack的workdir

**Features**

- ct-worker 镜像添加 cloudiac-playbook 命令

**Fixes**

- 修复漂移检测数组下标越界问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v1.2.0](https://github.com/idcos/cloudiac/releases/tag/v1.2.0)


------
## v1.1.0 20220819
**Features**

- 新增环境存活时间周期维度，通过crontab表达式设置环境定时部署，定时销毁

**Fixes**

- 修复修改环境配置时，环境标签被删除的问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v1.1.0](https://github.com/idcos/cloudiac/releases/tag/v1.1.0)


------
## v1.0.0 20220714
**Enhancements**

- 优化敏感变量展示，避免敏感信息泄露

**Features**

- 新增pipeline v0.5
- 新增平台概览
- 新增用户操作日志
- 新增用户找回密码功能
- 内置 terraform 增加 v1.1.9/v1.2.4 版本
- 新增Stack创建来源（Exchange）

**Fixes**

- 修复邮件通知重复发送问题
- 修复任务详情变量显示问题
- 修复屏蔽策略失败的问题
- 修复按策略屏蔽时数据查询异常问题
- 修复云模板列表中的repo地址展示异常问题
- 修复vcs分支/标签 默认只能返回20问题
- 修复解析合规检测结果失败问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v1.0.0](https://github.com/idcos/cloudiac/releases/tag/v1.0.0)


------
## v0.12.1 20220625
**Fixes**

- 修复项目审批员的访问权限问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.12.1](https://github.com/idcos/cloudiac/releases/tag/v0.12.1)


------
## v0.12.0 20220624
**Changes**

- 『云模板』统一更名为『Stack』

**Enhancements**

- 优化任务执行安全性，避免shell注入
- 优化注册密码强度
- 优化用户查看内容的权限
- 密钥管理支持设置RSA密钥

**Features**

- 新增资源查询依赖资源
- kafka回调消息增加任务id和合规状态字段

**Fixes**

- 修复plan后直接部署、销毁时 workdir 问题
- 修复趋势费用缺少2月份数据的问题
- 修复使用 Stack 导入功能，传入 json结构 时 panic 问题
- 修复环境标签只剩一个的情况下无法删除的问题
- 修复敏感变量加密问题
- 修复VCS相关的一些问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.12.0](https://github.com/idcos/cloudiac/releases/tag/v0.12.0)


------
## v0.11.0 20220530
**Enhancements**

- 优化环境可部署状态检查接口，现在只检查环境关联的云模板 和 vcs 是否有效
- 优化自动纠偏任务执行完 plan 后会判断是否有漂移，若无漂移则提前结束任务
- consul 服务创建锁并自动重新注册

**Features**

- 组织和项目支持批量邀请用户
- 新增，支持 ldap 登陆
- 新增，环境支持设置工作目录
- 新增，组织和项目概览页统计数据
- 新增，aliyun 资源费用采集
- 新增价格预估功能，在审批部署任务时展示资源变更的预估费用情况
- 增加 registry network mirror 支持，配置了 registry 服务地址后会自动启用该地址作为 network mirror server
- 接入自研 cloudcost 询价服务，目前支持的产品 aliyun ecs/disk/nat/slb/eip/rds/redis/mongodb

**Fixes**

- 修复设置工作目录后 tfvars 和 playbook 文件路径保存错误的问题
- 修复 playbook 中输出中文内容会乱码的问题
- 修复工作目录不支持二层以上子目录的问题
- 修复云模板中的敏感变量导入后变为乱码的问题
- 修复环境锁定后 plan 完成可以发起部署的问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.11.0](https://github.com/idcos/cloudiac/releases/tag/v0.11.0)


------
## v0.9.4 20220310
**Features**

- 任务结束后的 kafka 回调消息中增加任务类型和环境状态
- 销毁资源、重新部署接口增加 source 字段，第三方服务调用时可通过该字段设置触发来源



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.9.4](https://github.com/idcos/cloudiac/releases/tag/v0.9.4)


------
## v0.9.1 20220310
**Features**

- 环境支持设置及展示标签
- 环境创建、销毁、重新部署时都发送 kafka 消息，通知环境最新资源数据



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.9.1](https://github.com/idcos/cloudiac/releases/tag/v0.9.1)


------
## v0.9.0 20220307
**Enhancements**

- 优化 vcs 服务报错，将 vcs 错误进一步细分为连接错误、认证错误等
- 项目云模板列表的“活跃环境”字段改为关联环境，点击数字可跳转到环境列表页面

**Features**

- 合规策略组改用代码库进行管理，支持通过分支或 tag 来管理版本
- 增强合规检测引擎，细化云模板及环境检测流程
- 执行界面增加云模板和环境的合规开关和合规策略组绑定功能
- 新增合规管理员角色
- 新增环境搜索功能，支持通过环境名称和云模板名称进行搜索
- 环境部署历史增加触发类型字段，记录部署任务的触发来源

**Fixes**

- 修复设置工作目录后无法选择工作目录下的 ansible plabyook 和 tfvars 文件的问题
- 修复组织管理员无权修改项目名称和描述的问题
- 修复任务 plan 失败后可能长时间不退出的问题
- 修复 gitee 私有仓库无法认证的问题
- 修复导出云模板时资源账号的敏感变量加解密处理错误的问题
- 修复任务驳回后状态显示为“失败”的问题
- 修复触发器触发归档环境部署问题
- 修复部分查询未正常处理软删除的问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.9.0](https://github.com/idcos/cloudiac/releases/tag/v0.9.0)


------
## v0.8.1 20211214
**Fixes**

- 修复新组织中创建环境时接口报错的问题
- 修复环境有敏感变量时执行部署报解密错误的问题
- 修复执行任务容器异常退出会导致任务一直处于执行中状态且环境的资源一直累积的问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.8.1](https://github.com/idcos/cloudiac/releases/tag/v0.8.1)


------
## v0.8.0 20211210
**Enhancements**

- 优化环境列表和环境详情的展示效果

**Features**

- 新增环境漂移检测功能
- 新增环境资源、模型可视化展示
- 新增云模板及其关联数据的导出导入功能
- 新增创建云模板时名称和工作目录有效性检查
- 新增加 VCS 编辑功能
- 新增执行 MR/PR 触发的 plan 任务时将日志回写到 review comment
- 任务通知消息中增加任务类型说明

**Fixes**

- 修复步骤超时后不显示日志的问题
- 修复编辑云模板时仓库名称可能显示为 id 的问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.8.0](https://github.com/idcos/cloudiac/releases/tag/v0.8.0)


------
## v0.7.1 20211116
**Features**

- 新增 runner 的 offline mode

**Fixes**

- 修复预置 provider 不生效的问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.7.1](https://github.com/idcos/cloudiac/releases/tag/v0.7.1)


------
## v0.7.0 20211105
**Enhancements**

- 优化从组织和项目中移除用户功能
- 组织中编辑用户时允许修改姓名和手机号
- 优化环境列表、环境详情展示样式

**Features**

- 新增自定义 pipeline 功能，并将任务执行过程分步展示
- 新增组织内资源查询功能
- 新增资源账号管理功能

**Fixes**

- 修复从组织中删除用户后用户在项目中依然存在的问题
- 修复设置环境自动触发 plan/apply 功能报错的问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.7.0](https://github.com/idcos/cloudiac/releases/tag/v0.7.0)


------
## v0.6.0 20210928
**Features**

- 新增合规检测功能，平台管理员可进行合规策略管理
- 新增消息通知功能，支持邮件、钉钉、微信、Slack 事件通知
- 新增任务重试功能，可在环境设置中开启执行失败自动重试
- 新增 tfvars 文件和 playbook 文件内容查看功能
- 新增 terraform 版本选择功能，并支持版本的自动匹配
- 新增环境的资源详情展示，点击资源名称可查看资源详情
- 新增选择型变量，添加环境和 terraform 变量时可下拉选择
- 任务增加审批驳回状态，审批驳回不再显示为“失败”

**Fixed**

- 修复环境部署过程中允许删除关联云模板的问题
- 修复存在活跃环境的云模板在列表中活跃资源数显示为 0 的问题



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.6.0](https://github.com/idcos/cloudiac/releases/tag/v0.6.0)


------
## v0.5.1 20210806
**Features**

- 全新 0.5 版本发布



**完整 Changelog 及版本包:** [https://github.com/idcos/cloudiac/releases/tag/v0.5.1](https://github.com/idcos/cloudiac/releases/tag/v0.5.1)



