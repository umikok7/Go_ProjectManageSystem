package fs

import "os"

func IsExist(path string) bool {
	_, err := os.Stat(path) // 获取文件信息
	if err != nil {
		if os.IsNotExist(err) {
			return false // 文件不存在
		}
		return false
	}
	return true
}
