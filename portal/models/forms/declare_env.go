// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type DeclareEnvForm struct {
	BaseForm
	AppStack_  string `json:"app_stack"`
	AppStack   string `json:"appStack"`
	Cloud      string `json:"cloud"`
	Region     string `json:"region"`
	Zone       string `json:"zone"`
	ChargeType string `json:"charge_type"`

	Instances instance `json:"instances"`

	Recovery []recovery `json:"recovery"`

	ExtraData models.JSON `json:"extraData"`
}

type instance struct {
	InstanceNumber      string `json:"instanceNumber"`
	ChargeType          string `json:"chargeType"`
	InstanceUnit        string `json:"instanceUnit"`
	SysDiskCategory     string `json:"sysDiskCategory"`
	SysDiskPerformance  string `json:"sysDiskPerformance"`
	SysDiskSize         string `json:"sysDiskSize"`
	DataDiskSize        string `json:"dataDiskSize"`
	DataDiskCategory    string `json:"dataDiskCategory"`
	DataDiskPerformance string `json:"dataDiskPerformance"`
	InstanceType        string `json:"instanceType"`
	ImageId             string `json:"imageId"`
	InstanceChargeType  string `json:"instanceChargeType"`
	UserData            string `json:"userData"`
	Tags                string `json:"tags"`
	FirstIndex          string `json:"firstIndex"`
	EnvironmentId       string `json:"environmentId"`
	KeyName             string `json:"keyName"`
}

type recovery struct {
	Name       string `json:"name"`
	RecoveryId string `json:"recovery_id"`
}
