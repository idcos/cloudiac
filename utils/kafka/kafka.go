// Copyright 2021 CloudJ Company Limited. All rights reserved.

package kafka

import (
	"cloudiac/configs"
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
	TaskStatus string `json:"taskStatus" form:"taskStatus" `
}

type IacKafkaContent struct {
	TransactionId string                 `json:"transactionId"`
	Result        IacKafkaCallbackResult `json:"result"`
}

func (k *KafkaProducer) GenerateKafkaContent(transactionId, result string) []byte {
	a := IacKafkaContent{transactionId, IacKafkaCallbackResult{TaskStatus: result}}
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
	conf := sarama.NewConfig()
	conf.Producer.Retry.Max = 1
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
