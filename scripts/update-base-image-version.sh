#!/bin/bash

######################################################################################
# base image 版本更新脚本，该脚本读取 docker/base/VERSION 文件获取版本号，
# 然后修改所有组件的 Dockerfile，替换 FROM base-xxx:vxxx 中的 base image 版本号。
# 
# 使用方式:
#   1. 更新 docker/base/VERSION 内容为新的版本号
#   2. 在项目根目录下执行该脚本: bash scripts/update-base-image-version.sh
######################################################################################

MY_DIRNAME=$(dirname "$0")
MY_PATH=$(cd "${MY_DIRNAME}"; pwd)

OS_NAME="$(uname -s)"
VERSION=$(cat "$MY_PATH/../docker/base/VERSION") || exit $?

function sed_replace() {
  if [[ "$OS_NAME" == "Darwin" ]]; then
    sed -i '' "$@"
  else
    sed -i "$@"
  fi 
}

find "$MY_PATH/../docker" -name 'Dockerfile*' | \
  grep -v 'docker/base/' | while read FILE;do 
    sed_replace -re "s#^FROM cloudiac/(base-[^:]+):v[^-]+#FROM cloudiac/\1:$VERSION#g" "$FILE"
done

