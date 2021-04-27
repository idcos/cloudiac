package handlers

import (
	"fmt"
	"github.com/Shopify/sarama"
	"log"
)

type KafkaConfig struct {
	Brokers      []string `yaml:"brokers"`
	Topic        string   `yaml:"topic"`
	GroupID      string   `yaml:"group_id"`
	Partition    int      `yaml:"partition"`
	SaslUsername string   `yaml:"sasl_username"`
	SaslPassword string   `yaml:"sasl_password"`
	TLSCertFile  string   `yaml:"tls_cert_file"`
	TLSKeyFile   string   `yaml:"tls_key_file"`
	TLSCAFile    string   `yaml:"tls_ca_file"`
}

// KafkaProducer kafka 消息生产者
type KafkaProducer struct {
	Brokers   []string
	Topic     string
	Partition int32
	Conf      *sarama.Config
}

// ConnAndSend 连接并发送消息
func (k *KafkaProducer) ConnAndSend(msg []byte) (err error) {
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
	log.Println(fmt.Sprintf("KafkaProducer ConnAndSend send message success: %d %d", partition, offset))
	return nil
}


func InitKafkaProducerBuilder(kaConf KafkaConfig) *KafkaProducer {
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

	p := &KafkaProducer{
		Brokers:   kaConf.Brokers,
		Topic:     kaConf.Topic,
		Partition: int32(kaConf.Partition),
		Conf:      conf,
	}
	return p
}
