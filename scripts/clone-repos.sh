#!/bin/bash
##################################
## clone builtin repos

MY_DIRNAME=$(dirname "$0")
MY_PATH=$(cd "${MY_DIRNAME}"; pwd)

REPO_BASE=${REPO_BASE:-https://github.com/idcos}
REPOS_LIST=${REPOS_LIST:-$MY_PATH/../repos.list}

function clone() {
  local REPO_PATH=$1
  test -z "$REPO_PATH" && return 0

  local REPO_NAME=$(basename "${REPO_PATH}")
  local TARGET_DIR="$(dirname $MY_PATH)/repos/cloudiac/$REPO_NAME"

  if echo "$REPO_PATH" | grep '://' >/dev/null; then 
    local REPO_ADDRESS="${REPO_PATH}"
  else
    local REPO_ADDRESS="${REPO_BASE}/${REPO_PATH}"
  fi

  if [[ -d "$TARGET_DIR" ]]; then
    (set -x && cd "$TARGET_DIR" && git fetch) || return $?
  else
    (set -x && git clone --mirror "$REPO_ADDRESS" "$TARGET_DIR") || return $?
    cp "${TARGET_DIR}/hooks/post-update.sample" "${TARGET_DIR}/hooks/post-update"
  fi
  (cd "$TARGET_DIR" && bash hooks/post-update)
}

while read -r REPO_PATH; do
  echo "$REPO_PATH" | grep -E '^#' >/dev/null && continue
  clone "$REPO_PATH"
done < "$REPOS_LIST"

