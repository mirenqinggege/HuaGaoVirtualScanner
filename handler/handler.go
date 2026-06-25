// Package handler 消息处理器注册和分发
package handler

import (
	"encoding/json"
	"log"

	"github.com/huagao/virtual-scanner/scanner"
	"github.com/huagao/virtual-scanner/server"
	"github.com/huagao/virtual-scanner/store"
)

// NewRegistry 创建并注册所有消息处理器
func NewRegistry(st *store.Store, vs *scanner.VirtualScanner) *server.HandlerRegistry {
	reg := server.NewHandlerRegistry()

	// 二、基础功能接口
	registerBasicHandlers(reg, st, vs)

	// 三、图像采集接口
	registerDeviceHandlers(reg, st, vs)

	// 四、图像业务接口
	registerImageHandlers(reg, st, vs)

	log.Printf("✅ 已注册 %d 个消息处理器", reg.HandlerCount())

	return reg
}

// getString 从 raw message 中提取指定字段的字符串值
func getString(raw json.RawMessage, key string) string {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getBool 从 raw message 中提取指定字段的布尔值
func getBool(raw json.RawMessage, key string, defaultVal bool) bool {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return defaultVal
	}
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}

// getInt 从 raw message 中提取指定字段的整数值
func getInt(raw json.RawMessage, key string, defaultVal int) int {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return defaultVal
	}
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return defaultVal
}

// getFloat 从 raw message 中提取指定字段的浮点值
func getFloat(raw json.RawMessage, key string, defaultVal float64) float64 {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return defaultVal
	}
	if v, ok := m[key]; ok {
		switch f := v.(type) {
		case float64:
			return f
		case int:
			return float64(f)
		}
	}
	return defaultVal
}

// getStringSlice 从 raw message 中提取字符串数组
func getStringSlice(raw json.RawMessage, key string) []string {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	if v, ok := m[key]; ok {
		if arr, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

// getIntSlice 从 raw message 中提取整数数组
func getIntSlice(raw json.RawMessage, key string) []int {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	if v, ok := m[key]; ok {
		if arr, ok := v.([]interface{}); ok {
			result := make([]int, 0, len(arr))
			for _, item := range arr {
				switch n := item.(type) {
				case float64:
					result = append(result, int(n))
				case int:
					result = append(result, n)
				}
			}
			return result
		}
	}
	return nil
}
