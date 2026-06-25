// Package protocol 定义 WebSocket 通信的消息类型和序列化
package protocol

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

// Message 通用 JSON 消息结构
// 可以直接 Unmarshal 任意请求/响应/事件
type Message map[string]interface{}

// ParseRequest 从原始 JSON 字节解析请求
func ParseRequest(data []byte) (string, string, json.RawMessage, error) {
	var raw struct {
		Func string `json:"func"`
		Iden string `json:"iden"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", "", nil, fmt.Errorf("invalid request: %w", err)
	}
	if raw.Func == "" {
		return "", "", nil, fmt.Errorf("missing 'func' field")
	}
	// 返回完整的原始数据供具体 handler 复用
	return raw.Func, raw.Iden, data, nil
}

// SendResponse 发送响应消息
// extra 为 nil 或包含额外字段的 map，会合并到响应中
func SendResponse(conn *websocket.Conn, funcName string, iden string, ret int, errInfo string, extra map[string]interface{}) error {
	resp := Message{
		"func": funcName,
		"iden": iden,
		"ret":  ret,
	}
	if errInfo != "" {
		resp["err_info"] = errInfo
	}
	for k, v := range extra {
		resp[k] = v
	}
	return conn.WriteJSON(resp)
}

// SendOK 发送成功响应 (ret=0)
func SendOK(conn *websocket.Conn, funcName string, iden string, extra map[string]interface{}) error {
	return SendResponse(conn, funcName, iden, 0, "", extra)
}

// SendError 发送错误响应 (ret!=0)
func SendError(conn *websocket.Conn, funcName string, iden string, ret int, errInfo string) error {
	return SendResponse(conn, funcName, iden, ret, errInfo, nil)
}

// SendEvent 发送事件消息（不需要 ret 字段）
func SendEvent(conn *websocket.Conn, funcName string, iden string, extra map[string]interface{}) error {
	event := Message{
		"func": funcName,
		"iden": iden,
	}
	for k, v := range extra {
		event[k] = v
	}
	return conn.WriteJSON(event)
}
