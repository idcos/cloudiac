package kafka

import (
	"cloudiac/configs"
	"encoding/json"
	"fmt"
	"testing"
)

func TestKafkaProducer_ConnAndSend(t *testing.T) {
	configs.Init("../../config.yml")
	InitKafkaProducerBuilder()
	b, _ := json.Marshal("hello honghong")
	err := kafka.ConnAndSend(b)
	fmt.Println(err)
}
