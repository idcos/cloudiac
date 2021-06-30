package rabbitmq

import (
	"cloudiac/utils/logs"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type AMQP struct {
	mu           *sync.RWMutex
	conn         *amqp.Connection
	notifyClose  chan *amqp.Error
	disconnected bool
	Channels     chan *amqp.Channel
	Dial         func() (*amqp.Connection, error)
}

type ExchangeContext struct {
	ExchangeName string
	ExchangeKind string
	Queues       []*QueueContext
}

type QueueContext struct {
	QueueName  string
	RoutingKey string
	AutoAck    bool
}

var MQ *AMQP

// TimeAfter 超时函数
func TimeAfter(d time.Duration) chan int {

	q := make(chan int, 1)
	time.AfterFunc(d, func() {
		q <- 1
	})
	return q
}

// InitAMQP 初始化AMQP的连接
func InitAMQP(addr string) {
	if addr == "" {
		logs.Get().Warnf("'rabbitmq.addr' is empty, ignored")
		return
	}

	if MQ != nil {
		return
	}

	MQ = &AMQP{
		mu:           new(sync.RWMutex),
		disconnected: true,
		Channels:     make(chan *amqp.Channel, 100),
		notifyClose:  make(chan *amqp.Error),
		Dial: func() (*amqp.Connection, error) {
			return amqp.Dial(addr)
		},
	}
	MQ.reConnect()
}

func (mq *AMQP) reConnect() {
	var (
		err  error
		conn *amqp.Connection
	)
	// 重连5次, 每次间隔3秒
	for i := 0; i <= 5; i++ {
		conn, err = mq.Dial()
		if err == nil {
			fmt.Println("rabbitmq connect success")
			mq.conn = conn
			mq.disconnected = false

			// 连接关闭监听器
			go func() {
				errChan := make(chan *amqp.Error)
				for amqpErr := range conn.NotifyClose(errChan) {
					fmt.Println("rabbitmq disconnected %v, reconnecting", amqpErr)
					mq.disconnected = true
					mq.Channels = make(chan *amqp.Channel, 100)
					switch {
					case amqpErr.Code == 320:
						mq.reConnect()
					case amqpErr.Code == 501:
						mq.reConnect()
					case amqpErr.Code == 504:
						mq.reConnect()
					}
				}
			}()
			return
		}
		fmt.Println("rabbitmq connect failed", err)
		time.Sleep(3 * time.Second)
	}
}

// newChannel 基于底层的conn新建一个channel,多个channel共用一个conn
func (mq *AMQP) newChannel() (channel *amqp.Channel, err error) {
	if mq.disconnected {
		for i := 0; i < 10; i++ {
			if mq.disconnected {
				time.Sleep(3 * time.Second)
				fmt.Println("rabbitmq is disconnected waiting connect")
			} else {
				channel, err = mq.conn.Channel()
				return
			}
		}
		err = errors.New("rabbitmq disconnected")
		return
	}
	channel, err = mq.conn.Channel()
	return
}

// GetChannel 获取一个可用的channel
func (mq *AMQP) GetChannel() (channel *amqp.Channel, err error) {
	for {
		select {
		case channel = <-mq.Channels:
			fmt.Println("reusing channel")
			return
		case <-TimeAfter(time.Second * 1):
			fmt.Println("declare an new channel")
			channel, err = mq.newChannel()
			return
		}
	}
}

// ReleaseChannel 释放对应的channel
func (mq *AMQP) ReleaseChannel(channel *amqp.Channel) (closed bool) {
	defer func() {
		err := recover()
		if err != nil {
			mq.Channels = make(chan *amqp.Channel, 100)
			fmt.Println("release channel failed %v", err)
		}
	}()
	if !mq.disconnected {
		mq.Channels <- channel
	} else {
		channel.Close()
	}
	return
}

// DeclareExchange 创建对应的exchange
func (mq *AMQP) DeclareExchange(e *ExchangeContext) (err error) {
	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		return
	}
	// 名字、类型、是否持久化、是否在没有bind后自动删除、是否作为内部使用、是否要等待队列创建完成
	err = channel.ExchangeDeclare(e.ExchangeName, e.ExchangeKind, true, false, false, false, nil)
	if err != nil {
		fmt.Println("create exchange %s %v", e.ExchangeName, err)
	}
	return
}

// DeleteExchange 删除对应的exchange
func (mq *AMQP) DeleteExchange(e *ExchangeContext) (err error) {
	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		return
	}
	err = channel.ExchangeDelete(e.ExchangeName, false, false)
	return
}

// ExistsExchange 判断对应的exchange是否存在
func (mq *AMQP) ExistsExchange(e *ExchangeContext) (exists bool) {
	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		return
	}
	err = channel.ExchangeDeclarePassive(e.ExchangeName, e.ExchangeKind, true, false, false, false, nil)
	if err == nil {
		exists = true
	}
	return
}

// DeclareQueue 创建对应的Queue
func (mq *AMQP) DeclareQueue(q *QueueContext) (queue amqp.Queue, err error) {
	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		return
	}
	// 名字、是否持久化、是否自动删除、是否本次连接独占使用、是否等待创建结果
	queue, err = channel.QueueDeclare(q.QueueName, true, false, false, false, nil)
	return
}

// DeleteQueue 删除对应的Queue
func (mq *AMQP) DeleteQueue(q *QueueContext) (err error) {
	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		return
	}
	_, err = channel.QueueDelete(q.QueueName, false, false, false)
	return
}

// ExistsQueue 判断对应的Queue是否存在
func (mq *AMQP) ExistsQueue(q *QueueContext) (exists bool) {
	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		return
	}
	_, err = channel.QueueDeclarePassive(q.QueueName, true, false, false, false, nil)
	if err == nil {
		exists = true
	}
	return
}

// ExchangeBindWithQueue 将exchange和queue通过对应的routingkey绑定在一起
func (mq *AMQP) ExchangeBindWithQueue(exchangename, routingkey, queuename string) (err error) {
	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		return
	}
	err = channel.QueueBind(queuename, routingkey, exchangename, false, nil)
	return
}

//Send方法往某个消息队列发送消息
func (mq *AMQP) Send(queue string, body interface{}) {
	str, err := json.Marshal(body)
	if err != nil {
		fmt.Println("serialization message failed")
		return
	}

	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		fmt.Println("alloc channel failed %v", err)
		return
	}

	// 发送失败时，每隔1秒，重试三次
	loop := 0
	for {
		if err = channel.Publish(
			"",    //交换
			queue, //路由键
			false, //必填
			false, //立即
			amqp.Publishing{
				ReplyTo: queue,
				Body:    []byte(str),
			}); err != nil {
			if loop > 3 {
				return
			}
			loop++
			fmt.Println("publish message failed %s, retry %d", err, loop)
			time.Sleep(1 * time.Second)
		}
		break
	}
}

// Publish 向对应的exchange中发布消息
func (mq *AMQP) Publish(e *ExchangeContext, routingkey string, body string) (err error) {
	// 创建Exchange
	err = mq.DeclareExchange(e)
	if err != nil {
		fmt.Println("create exchange %s failed %v", e.ExchangeName, err)
		return
	}

	// 创建Queue并绑定
	for index := range e.Queues {
		qc := e.Queues[index]
		_, err = mq.DeclareQueue(qc)
		if err != nil {
			fmt.Println("create queue %s failed %v", qc.QueueName, err)
			return
		}
		err = mq.ExchangeBindWithQueue(e.ExchangeName, qc.RoutingKey, qc.QueueName)
		if err != nil {
			fmt.Println("bind exchange %s with queue %s failed %v", e.ExchangeName, qc.QueueName, err)
			return
		}
	}
	// 如何发生失败，每隔1秒，重试三次
	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		fmt.Println("alloc channel failed %v", err)
		return
	}

	loop := 0
	for {
		if err = channel.Publish(e.ExchangeName, routingkey, false, false, amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Persistent,
			Priority:        0,
		}); err != nil {
			if loop > 3 {
				return
			}
			loop++
			fmt.Println("publish message failed %s, retry %d", err, loop)
			time.Sleep(1 * time.Second)
		}
		break
	}
	return
}

// Subscribe 订阅对应queue中的消息
func (mq *AMQP) Subscribe(e *ExchangeContext, message chan []byte) (err error) {
	//// 创建Queue
	//_, err = mq.DeclareQueue(q)
	//if err != nil {
	//	log.Errorf("create queue %s failed %v", q.QueueName, err)
	//	return
	//}

	err = mq.DeclareExchange(e)
	if err != nil {
		fmt.Println("create exchange %s failed %v", e.ExchangeName, err)
		return
	}

	// 创建Queue并绑定
	for index := range e.Queues {
		qc := e.Queues[index]
		_, err = mq.DeclareQueue(qc)
		if err != nil {
			fmt.Println("create queue %s failed %v", qc.QueueName, err)
			return
		}
		err = mq.ExchangeBindWithQueue(e.ExchangeName, qc.RoutingKey, qc.QueueName)
		if err != nil {
			fmt.Println("bind exchange %s with queue %s failed %v", e.ExchangeName, qc.QueueName, err)
			return
		}
	}

	// 分配一个channel
	channel, err := mq.GetChannel()
	defer mq.ReleaseChannel(channel)
	if err != nil {
		fmt.Println("alloc channel failed %v", err)
		return
	}
	q := e.Queues[0]
	deliveries, err := channel.Consume(q.QueueName, "", q.AutoAck, false, false, false, nil)
	if err != nil {
		fmt.Println("rabbitmq consume failed, %v", err)
		return
	}
	go func(delivery <-chan amqp.Delivery, message chan []byte) {
		for d := range delivery {
			message <- d.Body
			// 回复本条消息的收到确认
			if !q.AutoAck {
				err := d.Ack(false)
				if err != nil {
					fmt.Println("ack received message failed %v", err)
				}
			}
		}
	}(deliveries, message)
	return
}
