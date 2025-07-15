package kk

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type LogData struct {
	Topic string
	// 需要接收json数据
	Data []byte
}

type KafkaWriter struct {
	w    *kafka.Writer
	data chan LogData
}

// GetWriter 生产者
func GetWriter(addr string) *KafkaWriter {
	w := &kafka.Writer{
		Addr:     kafka.TCP(addr),
		Balancer: &kafka.LeastBytes{},
	}
	k := &KafkaWriter{
		w:    w,
		data: make(chan LogData, 100),
	}
	go k.sendKafka() // go程开启监听
	return k
}

func (w *KafkaWriter) Send(data LogData) {
	w.data <- data
}

func (w *KafkaWriter) Close() {
	if w.w != nil {
		w.w.Close()
	}
}

func (w *KafkaWriter) sendKafka() {

	for {
		select {
		case data := <-w.data:
			messages := []kafka.Message{
				{
					Topic: data.Topic,
					Key:   []byte("logMsg"),
					Value: data.Data,
				},
			}

			var err error
			const retries = 3
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			for i := 0; i < retries; i++ {
				// attempt to create topic prior to publishing the message
				err = w.w.WriteMessages(ctx, messages...)
				if err == nil {
					break
				}
				if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
					time.Sleep(time.Millisecond * 250)
					continue
				}
				if err != nil {
					log.Printf("kafka send writeMessage error %s \n", err.Error())
				}
			}
		}
	}

}
