package config

import "test.com/project-common/kk"

var kw *kk.KafkaWriter

func InitKafkaWriter() func() {
	kw = kk.GetWriter("写配置文件里:9092")
	return kw.Close
}

func SendLog(data []byte) {
	kw.Send(kk.LogData{
		Topic: "msproject_data(写配置文件）",
		Data:  data,
	})
}
