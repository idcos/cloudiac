## 产品概述

CloudIaC是一个开源基础设施即代码管理平台，它以『环境即服务』的方式来管理基础设施以及应用所属的环境，提供远程运行环境来执行“部署”并简化Terraform、Ansible 等 IaC 框架的云部署治理。

![picture 1](./images/f2087a5c0f30e8d632e26aa55ebb4fa490566c919bf54a237b67afcf17c169b0.png)  

![picture 2](./images/65d05f9db39a4f6cfcb04f33c743e666ae934cdd86630c28daac9095368aec19.png)  

![picture 3](./images/d723f258264d7c915a36eb854ceb707eba051aa07d320a4f8a92f01d5103a777.png)  


将基础设施和应用以代码交付

通过Terraform和Ansible的结合，使用声明式的配置文件将基础设施和应用编写为代码，并使用VCS来管理并控制配置文件的版本。

![picture 4](./images/a22475b7ff53ee215aece0e8d7b5e2905afed4bba8d2c3b09c3918562840b3d0.png)  


在基础设施创建之前进行合规检测

CloudIaC融合OPA（Open Policy Agent）引擎，以策略即代码的方式对即将创建的基础设施进行合规检查，在安全风险和错误配置发生之前尽量降低它们。

![picture 5](./images/bc8b931fb223081dc580040f134618b99237ad8a6c817012dcd23615f86ac277.png)  

![picture 6](./images/339db904ba04cdd50000fb1d1aac5737fe61158e10c56a37e4a483f5d7046e2c.png)  


对环境进行漂移检测

开启环境的『漂移检测』，及时发现漂移的发生，接收漂移发生的通知并通过资源拓扑直观呈现漂移数据。

![picture 7](./images/6c1e87bbe2a3fc2be4b7c04e6a616a0990468b35af906c7a07480f04f7650608.png)  


公有云环境费用统计及预估

公有云环境月度费用统计，掌握云端环境资源整体费用及占比情况，同时可在环境执行计划预览时给出月度费用预估，供管理员审批时参考。
（设计稿）

CloudIaC Registry

CloudIaC提供私有Registry，管理Provider/Module/Policy，同时可作为Provider代理，解决企业私有化部署或网络不可达情况下Terraform使用困难的问题。

![picture 8](./images/4869cca5c7d57aaf32886c8dd7385cb34f07cdf907bba64c12235e44998731ed.png)  

![picture 10](./images/97ef2a49816cb9faa17380bb0aa4b77ddd26ebf5dd3c613291e8360052c77903.png)  
