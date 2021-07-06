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

    aliyunv1d124d3 = {
      source = "aliyun/alicloud"
      version = "1.124.3"
    }

    huaweicloudv1d24d2 = {
      source  = "huaweicloud/huaweicloud"
      version = "1.24.2"
    }

    vspherev1d26d0  = {
      source  = "hashicorp/vsphere"
      version = "1.26.0"
    }

    aliyunhashicorpv1d124d3 = {
      source = "hashicorp/alicloud"
      version = "1.124.3"
    }
  }
}
