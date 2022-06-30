<h1 align="center">CloudIaC</h1>
<h3 align="center">开源基础设施即代码环境管理平台</h3>
<p align="center">
  <a href="https://github.com/idcos/cloudiac"><img src="https://shields.io/github/license/idcos/cloudiac" alt="License: Apache-2.0"></a>
  <a href="https://idcos.github.io/cloudiac"><img src="https://readthedocs.org/projects/cloudiac/badge/?version=latest" alt="Docs"></a>
  <a href="https://github.com/idcos/cloudiac/releases"><img src="https://img.shields.io/github/v/release/idcos/cloudiac" alt="GitHub release"></a>
  <a href="https://github.com/idcos/cloudiac/releases/latest"><img src="https://img.shields.io/github/downloads/idcos/cloudiac/total" alt="Latest release"></a>
  <a href="https://github.com/idcos/cloudiac"><img src="https://img.shields.io/github/stars/idcos/cloudiac?color=%231890FF&style=flat-square" alt="Stars"></a>
</p>
<hr />

CloudIaC https://cloudiac.org 是基于基础设施即代码(IaC, Infrastructure as Code)构建的云环境自动化管理平台。

### CloudIaC 功能

-   **环境即服务**: 通过Terraform和Ansible的结合，以代码交付基础设施和应用，使用VCS来管理并控制代码的版本，一份代码对应一套或多套环境;
-   **安全合规**: 融合OPA（Open Policy Agent）引擎，以策略即代码的方式对即将创建的基础设施进行合规检查，在安全风险和错误配置发生之前尽量降低它们;
-   **漂移检测**: 及时发现配置漂移的发生，接收漂移发生的通知并通过资源拓扑直观呈现漂移数据;
-   **费用统计及预估**: 公有云环境月度费用统计，掌握云端环境资源整体费用及占比情况，同时可在资源创建及删除时给出费用预估;
-   **其它企业级特性**: 包含AD整合，作业管理，云账号管理，组织与租户隔离，环境锁定，审计合规，私有化部署等企业级功能;

### CloudIaC 优势

-   **全生命周期**：从环境资源的供给、应用自动部署、配置变更到环境销毁，覆盖整个资源及应用的完整生命周期；
-   **持续部署**：通过 Pipeline 无缝对接 CI 工具，将部署融入持续交付和 DevOps 体系；
-   **团队协作**：团队协作管理环境，支持不同管理层级角色授权。

### 文档

-   [完整文档](https://docs.cloudiac.org/)
-   [视频介绍](https://space.bilibili.com/2138433328/channel/seriesdetail?sid=1908688)

### Self-Hosted

- 准备一台 Linux 主机，安装 docker, docker-compose；
- 以 root 用户执行如下命令

```sh
curl -fsSL https://raw.githubusercontent.com/idcos/cloudiac-docs/master/script/cloudiac-docker.sh | bash
```

### CloudIaC in Cloud

- [**START FOR FREE**](https://app.cloudiac.org)
  - $0 per month
  - Unlimited Number of Users
  - Unlimited Number of Projects
  - Unlimited Access to Registry in [mainland](https://exchange.cloudiac.org)
  - Up to 5 Organizations
  - Commnity Support
  - Fully Integated CI/CD
  - 99.9% Guaranteed Uptime
  - Platform Security
  - Cloud Flexibility
  
- **Pro Team**(comming soon)
- **Enterprise**(comming soon)

### 社区

如果您在使用过程中有任何疑问或建议，欢迎提交 [GitHub Issue](https://github.com/idcos/cloudiac/issues/new/choose) 或加入我们的社区进一步交流沟通。

项目官网: https://cloudiac.org

#### 微信交流群
欢迎加入CloudIaC开源技术交流群：

微信群超过200人无法扫码进入，请添加CloudIaC助手为好友，助手将邀请您进群

<img src="https://user-images.githubusercontent.com/11749193/147626753-ca8069dc-3b6e-4989-ad7c-541ba97794ed.png" alt="助手二维码" width="200"/>


