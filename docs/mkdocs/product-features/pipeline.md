# Pipeline

CloudIaC 支持 Pipeline，将环境的 Plan、部署、销毁任务拆分为多个步骤，默认的我们为所有任务类型定义了标准步骤流程，用户可以自定义 pipeline，来增加步骤、调整步骤执行顺序等。

同时在前端，任务也是分步骤展示： ![img](https://cloudiac.readthedocs.io/zh/latest/resources/pipeline-steps-1.png)

## 如何使用 Pipeline

自定义 Pipeline 通过在云模板代码库中增加 pipeline 文件实现，文件名为 `.cloudiac-pipeline.yml`。

文件查找逻辑如下:

1. 如果云模板设置了工作目录，则在工作目录下查找 `.cloudiac-pipeline.yml`，存在则使用
2. 如果工作目录不存在则在代码库根目录下查找 `.cloudiac-pipeline.yml`，存在则使用
3. 否则使用默认的 pipeline 标准流程模板

## Pipeline 标准流程模板

Pipeline 使用 yaml 格式定义，标准的流程模板如下:

```yaml
version: 0.3    # pipeline 格式版本号，为了保证多版本的兼容性

# plan 任务
plan:
  steps:    # 任务的步骤
    - type: checkout    # 步骤类型
      name: Checkout Code   # 步骤的展示名称，未提供名称则展示为步骤类型

    - type: terraformInit
      name: Terraform Init

    - type: opaScan
      name: OPA Scan

    - type: terraformPlan
      name: Terraform Plan

apply:
  steps:
    - type: checkout
      name: Checkout Code

    - type: terraformInit
      name: Terraform Init

    - type: opaScan
      name: OPA Scan

    - type: terraformPlan
      name: Terraform Plan

    - type: terraformApply
      name: Terraform Apply

    - type: ansiblePlay
      name: Run playbook

destroy:
  steps:
    - type: checkout
      name: Checkout Code

    - type: terraformInit
      name: Terraform Init

    - type: terraformPlan
      name: Terraform Plan
      args: 
        - "-destroy"

    - type: terraformDestroy
      name: Terraform Destroy
```

## Pipeline 的任务类型和步骤

从标准 pipeline 模板中可以看到，CloudIaC 支持对 plan、apply、destroy 三种任务进行自定义，分别对应环境的 Plan、部署任务、销毁任务。

自定义 pipeline 时可以只对指定的任务类型做定义，未定义的任务类型会使用标准 pipeline 流程步骤。

每一个任务类型都可以定义各自的步骤列表，CloudIaC 支持的步骤列表如下:

| 步骤类型         | 说明                      |
| ---------------- | ------------------------- |
| checkout         | 代码检出                  |
| terraformInit    | terraform init            |
| terraformPlan    | terraform plan            |
| terraformApply   | terraform apply           |
| terraformDestroy | terraform destroy         |
| ansiblePlay      | ansible-playbook          |
| opaScan          | OPA 策略扫描              |
| regoParse        | 解析资源属性为 rego input |
| command          | 执行自定义命令            |

同时步骤还支持 args 参数，terraform 和 ansible 相关步骤类型的 args 会以命令行参数的形式传递给执行的命令，如 terraformPlan 步骤传入 "-destroy" 参数用于生成 terraform destroy，command 步骤的 args 参数表示需要执行的 shell 命令。

## Command 步骤类型

command 步骤允许您执行任意 shell 命令，基于 command 命令您可以实现功能强大的自定义流程。

一些 command 命令使用场景示例:

```yaml
apply:
  steps:
    - name: Instal amazon.aws
      type: command
      args: 
        - yum install -y python2-pip
        - pip install botocore==1.21.41 boto3==1.18.41
        - ansible-galaxy collection install amazon.aws

    - name: Download alicloud provider
      type: command
      args: 
        - "curl -Ls https://github.com/aliyun/terraform-provider-alicloud/releases/download/v1.126.0/terraform-provider-alicloud_1.126.0_linux_amd64.zip >/usr/share/terraform/plugins/registry.terraform.io/aliyun/alicloud/terraform-provider-alicloud_1.126.0_linux_amd64.zip"

    - name: Trigger workflow
      type: command
      args: 
        - "curl -d token=${WORKFLOW_TOKEN} https://workflow.example.com/step/${WORKFLOW_STEPID}/start"
```

- 为避免与 yaml 格式特殊字符冲突，args 参数建议使用双引号包含
- CloudIaC 的任务步骤都是在容器中执行，不会影响宿主系统

## Pipeline 回调

除了给任务定义步骤之后 CloudIaC 还支持定义回调步骤，回调步骤基于任务的运行状态选择性执行。目前支持的回调类型有 `onSuccces` 和 `onFail`，onSuccess 步骤在任务所有步骤执行成功时回调，onFail 步骤在任务任意步骤执行失败时回调。

回调使用示例:

```yaml
apply:
  onSuccess:
    name: 任务成功
    type: command
    args: 
      - echo "Task successful"
      - test "$CLOUDIAC_ENV_STATUS" = "inactive" && echo "Environment created"
      - test "$CLOUDIAC_ENV_STATUS" = "failed" && echo "Environment recovered"

  onFail:
    name: 任务失败
    type: command
    args: 
      - echo "Task failed"

  steps: [] # 流程步骤省略
```

在 command 步骤中可以引用环境变量的值，通过判断变量值来执行不同的操作。如示例中基于环境在任务启动时的状态来判断是创建环境还是恢复失败状态的环境。

CLOUDIAC_ENV_STATUS 为任务启动时平台自动导出的环境变量，完整的导出环境变量列表见文档: [变量](https://cloudiac.readthedocs.io/zh/latest/intra/variable/)

回调步骤总是在流程的最后展示，流程步骤展示效果: ![img](https://cloudiac.readthedocs.io/zh/latest/resources/pipeline-steps-2.png)

## 完整的自定义 Pipeline 示例

一个完整的自定义 pipeline 示例：

```yaml
--- 
version: 0.3

plan:
  steps: 
    - type: checkout
      name: "Checkout code"

    - type: terraformInit
      name: "Terraform Init"

    - type: opaScan
      name: OPA Scan

    - type: terraformPlan
      name: "Terraform Plan"


apply: 
  onFail: 
    type: command
    args: 
      - echo "Job failed"

  onSuccess: 
    type: command
    args: 
      - echo "Job successful"

  steps: 
    - type: command
      args: 
        - "echo \"get somethings\""
        - "bash script.sh"
        - "curl 127.0.0.1/api/action"

    - type: checkout
      name: "Checkout code"

    - type: terraformInit
      name: "Terraform Init"

    - type: opaScan
      name: OPA Scan

    - type: terraformPlan
      name: "Terraform Plan"

    - type: terraformApply
      name: "Terraform Apply"

    - type: ansiblePlay
      name: "Run playbook"



destroy: 
  steps: 
    - type: checkout
      name: "Checkout code"

    - type: terraformInit
      name: "Terraform Init"

    - type: terraformPlan
      name: "Terraform Plan"
      args: 
        - "-destroy"

    - type: terraformDestroy
      name: "Terraform Destroy"

    - type: command
      name: "Say Bye"
      args: 
        - echo "Bye!"
```
