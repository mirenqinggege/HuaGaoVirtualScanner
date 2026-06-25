package scanner

import "encoding/base64"

// encodeBase64 将字节数据编码为 base64 字符串
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
