// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package task_manager

import (
	"cloudiac/portal/models"
	"testing"
)

var TextCase = `
random_password.password[0]: Refreshing state... [id=none]
random_password.password[1]: Refreshing state... [id=none]

Terraform used the selected providers to generate the following execution
plan. Resource actions are indicated with the following symbols:
  + create
-/+ destroy and then create replacement

Terraform will perform the following actions:

  # random_password.password[0] must be replaced
-/+ resource "random_password" "password" {
      ~ id               = "none" -> (known after apply)
      ~ length           = 12 -> 13 # forces replacement
      ~ result           = (sensitive value)
        # (9 unchanged attributes hidden)
    }

  # random_password.password[1] must be replaced
-/+ resource "random_password" "password" {
      ~ id               = "none" -> (known after apply)
      ~ length           = 12 -> 13 # forces replacement
      ~ result           = (sensitive value)
        # (9 unchanged attributes hidden)
    }

  # random_password.password[2] will be created
  + resource "random_password" "password" {
      + id               = (known after apply)
      + length           = 13
      + lower            = true
      + min_lower        = 0
      + min_numeric      = 0
      + min_special      = 0
      + min_upper        = 0
      + number           = true
      + override_special = "_%@"
      + result           = (sensitive value)
      + special          = true
      + upper            = true
    }

Plan: 3 to add, 0 to change, 2 to destroy.

─────────────────────────────────────────────────────────────────────────────

Note: You didn't use the -out option to save this plan, so Terraform can't
guarantee to take exactly these actions if you run "terraform apply" now.
`

func TestParseResourceDriftInfo(t *testing.T) {
	DriftMap := map[string]models.ResourceDrift{
		"random_password.password[0]": {
			DriftDetail: `-/+ resource "random_password" "password" {
      ~ id               = "none" -> (known after apply)
      ~ length           = 12 -> 13 # forces replacement
      ~ result           = (sensitive value)
        # (9 unchanged attributes hidden)
    }`},
		"random_password.password[1]": {
			DriftDetail: `-/+ resource "random_password" "password" {
      ~ id               = "none" -> (known after apply)
      ~ length           = 12 -> 13 # forces replacement
      ~ result           = (sensitive value)
        # (9 unchanged attributes hidden)
    }`},
		"random_password.password[2]": {
			DriftDetail: `  + resource "random_password" "password" {
      + id               = (known after apply)
      + length           = 13
      + lower            = true
      + min_lower        = 0
      + min_numeric      = 0
      + min_special      = 0
      + min_upper        = 0
      + number           = true
      + override_special = "_%@"
      + result           = (sensitive value)
      + special          = true
      + upper            = true
    }`},
	}
	tt := ParseResourceDriftInfo([]byte(TextCase))
	if tt["random_password.password[0]"].DriftDetail != DriftMap["random_password.password[0]"].DriftDetail {
		t.Error("random_password.password[0]解析失败") //nolint
	}
	if tt["random_password.password[1]"].DriftDetail != DriftMap["random_password.password[1]"].DriftDetail {
		t.Error("random_password.password[1]解析失败") //nolint
	}
	if tt["random_password.password[2]"].DriftDetail != DriftMap["random_password.password[2]"].DriftDetail {
		t.Error("random_password.password[2]解析失败") //nolint
	}

}
