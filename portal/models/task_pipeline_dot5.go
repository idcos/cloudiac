// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

const pipelineV0dot5 = `
version: 0.5

plan:
  onSuccess:
    type: string
    name: string
    args:
      - type: string

  onFail:
    type: string
    name: string
    args:
      - type: string

  steps:
    checkout:
      name: Checkout Code
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

    terraformInit:
      name: Terraform Init
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

    terraformPlan:
      name: Terraform Plan
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

    envScan:
      name: OPA Scan
      timeout: int

apply:
  onSuccess:
    type: string
    name: string
    args:
      - type: string

  onFail:
    type: string
    name: string
    args:
      - type: string

  steps:
    checkout:
      name: Checkout Code
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

    terraformInit:
      name: Terraform Init
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

    terraformPlan:
      name: Terraform Plan
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

    envScan:
      name: OPA Scan
      timeout: int

    terraformApply:
      name: Terraform Apply
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

    ansiblePlay:
      name: Run playbook
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

destroy:
  onSuccess:
    type: string
    name: string
    args:
      - type: string

  onFail:
    type: string
    name: string
    args:
      - type: string

  steps:
    checkout:
      name: Checkout Code
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

    terraformInit:
      name: Terraform Init
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

    terraformPlan:
      name: Terraform Plan
      timeout: int
      args:
        - "-destroy"
      before:
        - type: string
      after:
        - type: string

    envScan:
      name: OPA Scan
      timeout: int

    terraformDestroy:
      name: Terraform Apply
      timeout: int
      args:
        - type: string
      before:
        - type: string
      after:
        - type: string

# scan 和 parse 暂不开发自定义工作流
envScan:
  steps:
    - type: checkout
    - type: terraformInit
    - type: terraformPlan
    - type: envScan

envParse:
  steps:
    - type: checkout
    - type: terraformInit
    - type: terraformPlan
    - type: envParse

tplScan:
  steps:
    - type: scaninit
    - type: tplScan

tplParse:
  steps:
    - type: scaninit
    - type: tplParse
`
