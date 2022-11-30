// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package kafka

import (
	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
)

// KafkaProducer kafka 消息生产者
type KafkaProducer struct {
	Brokers   []string
	Topic     string
	Partition int32
	Conf      *sarama.Config
}

var kafka *KafkaProducer

type IacKafkaCallbackResult struct {
	Resources      []models.Resource               `json:"resources"`
	Outputs        map[string]interface{}          `json:"outputs"`
	DriftResources map[string]models.ResourceDrift `json:"drift_resources"`
}

type IacKafkaContent struct {
	EventType    string                 `json:"eventType"` // 当发生漂移后 固定为 task:drift_detection
	TaskStatus   string                 `json:"taskStatus"`
	TaskType     string                 `json:"taskType"`
	EnvStatus    string                 `json:"envStatus"`
	OrgId        models.Id              `json:"orgId"`
	ProjectId    models.Id              `json:"projectId"`
	TplId        models.Id              `json:"tplId"`
	EnvId        models.Id              `json:"envId"`
	TaskId       models.Id              `json:"taskId"`
	ExtraData    interface{}            `json:"extraData"`
	PolicyStatus string                 `json:"policyStatus"`
	IsDrift      bool                   `json:"isDrift"` // 漂移状态
	Result       IacKafkaCallbackResult `json:"result"`
}

func (k *KafkaProducer) GenerateKafkaDriftContent(task *models.Task, envStatus string,
	isDrift bool, driftResources map[string]models.ResourceDrift) []byte {

	a := IacKafkaContent{
		EventType:  common.DriftEventType,
		TaskStatus: task.Status,
		TaskType:   task.Type,
		EnvStatus:  envStatus,
		OrgId:      task.OrgId,
		ProjectId:  task.ProjectId,
		TplId:      task.TplId,
		EnvId:      task.EnvId,
		TaskId:     task.Id,
		IsDrift:    isDrift,
		Result: IacKafkaCallbackResult{
			DriftResources: driftResources,
		},
	}

	rep, _ := json.Marshal(&a)
	return rep
}

func (k *KafkaProducer) GenerateKafkaContent(task *models.Task, taskStatus, envStatus, policyStatus string,
	resources []models.Resource, outputs map[string]interface{}) []byte {
	a := IacKafkaContent{
		TaskStatus:   taskStatus,
		TaskType:     task.Type,
		EnvStatus:    envStatus,
		OrgId:        task.OrgId,
		ProjectId:    task.ProjectId,
		TplId:        task.TplId,
		EnvId:        task.EnvId,
		TaskId:       task.Id,
		IsDrift:      common.IsNotDrift,
		PolicyStatus: policyStatus,
		Result: IacKafkaCallbackResult{
			Resources: resources,
			Outputs:   outputs,
		},
	}

	if task.ExtraData != nil {
		a.ExtraData = task.ExtraData
	} else {
		a.ExtraData = make(map[string]interface{})
	}

	rep, _ := json.Marshal(&a)
	return rep
}

// ConnAndSend 连接并发送消息
func (k *KafkaProducer) ConnAndSend(msg []byte) (err error) {
	logger := logs.Get().WithField("kafka", "SendResultToKafka")
	syncProducer, err := sarama.NewSyncProducer(k.Brokers, k.Conf)
	if err != nil {
		return err
	}
	partition, offset, err := syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic:     k.Topic,
		Partition: k.Partition,
		Value:     sarama.ByteEncoder(msg),
	})
	if err != nil {
		return err
	}
	_ = syncProducer.Close()
	logger.Info(fmt.Sprintf("KafkaProducer ConnAndSend send message success: %d %d", partition, offset))
	return nil
}

func InitKafkaProducerBuilder() {
	kaConf := configs.Get().Kafka
	if kaConf.Disabled || len(kaConf.Brokers) <= 0 {
		logs.Get().Info("kafka was not open")
		return
	}

	conf := sarama.NewConfig()
	conf.Producer.Retry.Max = 3
	conf.Producer.RequiredAcks = sarama.WaitForLocal
	conf.Producer.Return.Successes = true
	conf.Metadata.Full = true
	conf.Version = sarama.V2_5_0_0
	conf.Consumer.Offsets.AutoCommit.Enable = true

	if len(kaConf.SaslUsername) > 0 {
		conf.Net.SASL.Enable = true
		conf.Net.SASL.User = kaConf.SaslUsername
		conf.Net.SASL.Password = kaConf.SaslPassword
		conf.Net.SASL.Handshake = true
	}

	kafka = &KafkaProducer{
		Brokers:   kaConf.Brokers,
		Topic:     kaConf.Topic,
		Partition: int32(kaConf.Partition),
		Conf:      conf,
	}
}

func Get() *KafkaProducer {
	if kafka == nil {
		InitKafkaProducerBuilder()
	}
	return kafka
}
