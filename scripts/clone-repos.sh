#!/bin/bash
##################################
## clone repos

MY_DIRNAME=$(dirname "$0")
MY_PATH=$(cd "${MY_DIRNAME}"; pwd)

REPO_BASE=${REPO_BASE:-https://github.com/cloud-iac/}
REPOS_LIST=${REPOS_LIST:-$MY_PATH/../repos.list}

function clone() {
  local REPO_PATH=$1
  local REPO_NAME=$(basename "${REPO_PATH}")
  local REPO_ADDRESS="${REPO_BASE}${REPO_PATH}"
  if [[ -d "$REPO_PATH" ]]; then
    git fetch "$REPO_ADDRESS" || return $?
  else
    git clone --bare "$REPO_ADDRESS" || return $?
    cp "$REPO_NAME"/hooks/post-update.sample "$REPO_NAME"/hooks/post-update
  fi
  (cd "$REPO_NAME" && bash hooks/post-update)
}

while read -r REPO_PATH; do
  clone "$REPO_PATH"
done < "$REPOS_LIST"
