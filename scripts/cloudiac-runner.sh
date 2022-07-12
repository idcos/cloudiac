#!/bin/bash

# 用法说明：
# 1. 将 cloudiac-runner.sh 和 terraform.py 放到 stack 目录下
# 2. 环境变量放在 .env 文件，terraform 变量放在 xx.tfvars 文件，
# 编辑 cloudiac-runner.sh 修改 TF_VAR_FILE 指向 tfvars 文件。
# 3. 如果使用了 ansible 部署，需要配置 ssh 密钥，私钥编辑
# cloudiac-runner.sh 修改 PRIVATE_KEY_FILE 指向私钥文件。
# 4. 执行：
#     部署: ./cloudiac-runner.sh
#     销毁: ./cloudiac-runner.sh destroy

# 参数解析
if [ "$1" == "destroy" ]; then
  OP_DESTROY=1
fi

# 执行参数设置
TF_VAR_FILE="./qa-env.tfvars"
PRIVATE_KEY_FILE="~/.ssh/id_rsa"
PLAYBOOK_FILE="./ansible/playbook.yml"

# 导入环境变量
set -o allexport; source .env; set +o allexport

# 安装 terraform
terraform -v >/dev/null 2>&1
if [ $? -ne 0 ]; then
  sudo yum install -y yum-utils && sudo yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo && sudo yum -y install terraform
  if [ $? -ne 0 ]; then
    exit
  fi
fi

# 安装 ansible
ansible --version >/dev/null 2>&1
if [ $? -ne 0 ]; then
  sudo yum install epel-release -y && sudo yum install ansible -y
  if [ $? -ne 0 ]; then
    exit
  fi
fi

# 配置 provider 镜像
if [ ! -f "~/.terraformrc" ]; then
  cat << EOF > ~/.terraformrc
provider_installation {
  network_mirror {
    url = "https://exchange.cloudiac.org/v1/mirrors/providers/"
    include = ["registry.terraform.io/*/*"]
  }

  direct {
    exclude = ["registry.terraform.io/*/*"]
  }
}
EOF
fi

# 部署/销毁资源
if [ ! $OP_DESTROY ]; then
  terraform init && terraform apply -var-file=$TF_VAR_FILE
  if [ $? -ne 0 ]; then
    exit
  fi
else
  terraform init && terraform destroy -var-file=$TF_VAR_FILE
  if [ $? -ne 0 ]; then
    exit
  fi
  exit
fi

# 执行 ansible playbook
export ANSIBLE_HOST_KEY_CHECKING="False"
export ANSIBLE_TF_DIR="."
export ANSIBLE_NOCOWS="1"

ansible-playbook \
	--inventory terraform.py \
	--user "root" \
	--private-key $PRIVATE_KEY_FILE \
	$PLAYBOOK_FILE
