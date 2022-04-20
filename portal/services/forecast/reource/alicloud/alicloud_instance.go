package alicloud

import (
	"cloudiac/portal/services/forecast/schema"
	"fmt"
)

type Instance struct {
	Address                        string
	Region                         string
	AllocatePublicIp               interface{}   `json:"allocate_public_ip"`
	AutoReleaseTime                interface{}   `json:"auto_release_time"`
	AutoRenewPeriod                interface{}   `json:"auto_renew_period"`
	AvailabilityZone               string        `json:"availability_zone"`
	DataDisks                      []interface{} `json:"data_disks"`
	DeletionProtection             bool          `json:"deletion_protection"`
	DeploymentSetId                interface{}   `json:"deployment_set_id"`
	Description                    interface{}   `json:"description"`
	DryRun                         bool          `json:"dry_run"`
	ForceDelete                    interface{}   `json:"force_delete"`
	HpcClusterId                   interface{}   `json:"hpc_cluster_id"`
	ImageId                        string        `json:"image_id"`
	IncludeDataDisks               interface{}   `json:"include_data_disks"`
	InstanceChargeType             string        `json:"instance_charge_type"`
	InstanceName                   string        `json:"instance_name"`
	InstanceType                   string        `json:"instance_type"`
	InternetChargeType             string        `json:"internet_charge_type"`
	InternetMaxBandwidthOut        int           `json:"internet_max_bandwidth_out"`
	IoOptimized                    interface{}   `json:"io_optimized"`
	IsOutdated                     interface{}   `json:"is_outdated"`
	KmsEncryptedPassword           interface{}   `json:"kms_encrypted_password"`
	KmsEncryptionContext           interface{}   `json:"kms_encryption_context"`
	Password                       interface{}   `json:"password"`
	Period                         interface{}   `json:"period"`
	PeriodUnit                     interface{}   `json:"period_unit"`
	RenewalStatus                  interface{}   `json:"renewal_status"`
	ResourceGroupId                interface{}   `json:"resource_group_id"`
	SecurityEnhancementStrategy    interface{}   `json:"security_enhancement_strategy"`
	SpotPriceLimit                 interface{}   `json:"spot_price_limit"`
	SpotStrategy                   string        `json:"spot_strategy"`
	Status                         string        `json:"status"`
	SystemDiskAutoSnapshotPolicyId interface{}   `json:"system_disk_auto_snapshot_policy_id"`
	SystemDiskCategory             string        `json:"system_disk_category"`
	SystemDiskDescription          interface{}   `json:"system_disk_description"`
	SystemDiskName                 interface{}   `json:"system_disk_name"`
	SystemDiskSize                 int64         `json:"system_disk_size"`
	Tags                           interface{}   `json:"tags"`
	Timeouts                       interface{}   `json:"timeouts"`
	UserData                       interface{}   `json:"user_data"`
}

func (a *Instance) BuildResource() *schema.Resource {
	p := make([]*schema.PriceRequest, 0)

	if a.InstanceType != "" {
		p = append(p, &schema.PriceRequest{
			Name:  "InstanceType",
			Value: fmt.Sprintf("InstanceType:%s", a.InstanceType),
		})
	}

	if a.SystemDiskSize != 0 && a.SystemDiskCategory != "" {
		p = append(p, &schema.PriceRequest{
			Name:  "SystemDisk",
			Value: fmt.Sprintf("SystemDisk.Category:%s,SystemDisk.Size:%d", a.SystemDiskCategory, a.SystemDiskSize),
		})
	}

	return &schema.Resource{
		Name:        a.Address,
		RequestData: p,
		PriceCode:   "ecs",
	}
}
