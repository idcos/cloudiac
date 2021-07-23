#!/bin/bash
##############################################################
# 下载所有内置代码库中依赖的 providers 到 TARGET_DIR 目录

## 目标目录
TARGET_DIR=${TARGET_DIR:-./assets/providers}

## 内置代码库的路径
REPOS=${REPOS:-./repos}

## 指定 terraform 插件的平台和架构
PLATFORM=${PLATFORM:-linux_amd64}

MY_DIRNAME=$(dirname "$0")
MY_PATH=$(cd "${MY_DIRNAME}"; pwd)
## providers.tf 文件路径，用于额外下载指定的 provider
PROVIDERS_FILE_DIR=${PROVIDERS_FILE_DIR:-$MY_PATH}

function terraform_mirror() {
    local WORKDIR=$1
    local TARGET_DIR=$(cd "${TARGET_DIR}"; pwd)  # 使用绝对路径
    ## 在子 shell中处理，保持当前 shell 路径不变
    (cd "${WORKDIR}" && terraform providers mirror -platform="${PLATFORM}" "${TARGET_DIR}")
}
   
mkdir -p "$TARGET_DIR"
terraform_mirror "${PROVIDERS_FILE_DIR}"

find "${REPOS}" -name '*?.git' | while read -r REPO; do
    git clone "${REPO}" "${REPO}/tree" >/dev/null && \
      ls "${REPO}"/tree/*.tf >/dev/null && terraform_mirror "${REPO}/tree"

    ## 确保执行完成后删除 clone 出来的 tree 目录
    test -d "${REPO}/tree" && rm -rf "${REPO}/tree"
done

