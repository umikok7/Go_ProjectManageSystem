package kk

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
)

type KafkaReader struct {
	r *kafka.Reader
}

// readMsg 消费者
func (r *KafkaReader) readMsg() {
	for {
		m, err := r.r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("kafka read message error %s \n", err.Error())
			continue
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v： %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}
}

func GetReader(broker []string, groupId, topic string) *KafkaReader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  broker,
		GroupID:  groupId, // 同一个组下的consumer协同工作，共同消费topic队列中的内容
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	k := &KafkaReader{
		r: r,
	}
	go k.readMsg()
	return k
}
