// 这里指定的 provider 会预下载并打包到 providers 目录, 参考: generate_providers_mirror.sh

terraform {
  required_providers {
    ansible = {
      source = "nbering/ansible"
    }

    cloudinitvlatest = {
      source = "hashicorp/cloudinit"
    }

    aliyunvlatest = {
      source = "aliyun/alicloud"
    }

    huaweicloudv1dot24dot2 = {
      source  = "huaweicloud/huaweicloud"
      version = "1.24.2"
    }
  }
}
