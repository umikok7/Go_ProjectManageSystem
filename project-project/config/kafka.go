package config

import (
	"test.com/project-common/kk"
)

var kw *kk.KafkaWriter

func InitKafkaWriter() func() {
	kw = kk.GetWriter(C.KafkaConfig.Addr)
	return kw.Close
}

func SendLog(data []byte) {
	kw.Send(kk.LogData{
		Topic: C.KafkaConfig.Topic,
		Data:  data,
	})
}
