// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package kafka

import (
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
	Resources []models.Resource `json:"resources"  `
}

type IacKafkaContent struct {
	EventType  string                 `json:"eventType"`
	ExtraData  models.JSON            `json:"extraData"`
	TaskStatus string                 `json:"taskStatus"`
	OrgId      models.Id              `json:"orgId"`
	ProjectId  models.Id              `json:"projectId"`
	TplId      models.Id              `json:"tplId"`
	EnvId      models.Id              `json:"envId"`
	Result     IacKafkaCallbackResult `json:"result"`
}

func (k *KafkaProducer) GenerateKafkaContent(task *models.Task, taskStatus string, resources []models.Resource) []byte {
	a := IacKafkaContent{
		ExtraData:  task.ExtraData,
		TaskStatus: taskStatus,
		OrgId:      task.OrgId,
		ProjectId:  task.ProjectId,
		TplId:      task.TplId,
		EnvId:      task.EnvId,
		Result: IacKafkaCallbackResult{
			Resources: resources,
		},
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
	if kaConf.Disabled {
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
