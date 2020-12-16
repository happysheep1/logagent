package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"time"
)

type logData struct {
	topic string
	data  string
}

//往kafka写日志的模块
var (
	client      sarama.SyncProducer //声明一个连接kafka的全局生产者client
	logDataChan chan *logData
)

func Init(addrs []string, maxSize int) (err error) {
	config := sarama.NewConfig()
	//tailf包使用
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出⼀一个partition ，NewRandomPartitioner 轮询
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回
	// 连接kafka
	client, err = sarama.NewSyncProducer(addrs, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return
	}
	//为了可以动态配置管道的长度
	logDataChan = make(chan *logData, maxSize)
	//开启后台goroutine从通道中取数据发往kafka
	go Send2Kafka()
	return nil
}

//真正往kafka发送消息的地方
func Send2Kafka() {
	for {
		select {
		case lg := <-logDataChan:
			msg := &sarama.ProducerMessage{}
			msg.Topic = lg.topic
			msg.Value = sarama.StringEncoder(lg.data)
			// 发送消息
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("send msg failed, err:", err)
				return
			}
			fmt.Printf("pid:%v offset:%v\n", pid, offset)
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}

}

//给外部暴露一个函数，这个函数的作用是把日志数据发送到一个内部的chan中
func Send2Chan(topic, data string) {
	msg := &logData{
		topic: topic,
		data:  data,
	}
	logDataChan <- msg
}
