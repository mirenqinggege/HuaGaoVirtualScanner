// Package server 提供 WebSocket 服务器、连接管理和消息路由
package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/huagao/virtual-scanner/protocol"
	"github.com/huagao/virtual-scanner/store"
)

// Session 代表一个客户端连接
type Session struct {
	Conn   *websocket.Conn
	Store  *store.Store
	Handler *HandlerRegistry

	mu      sync.Mutex
	closed  bool
	sendCh  chan interface{} // 发送队列
	closeCh chan struct{}
}

// HandlerFunc 消息处理函数类型
type HandlerFunc func(session *Session, iden string, raw json.RawMessage)

// HandlerRegistry 消息处理器注册表
type HandlerRegistry struct {
	handlers map[string]HandlerFunc
}

// NewHandlerRegistry 创建处理器注册表
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]HandlerFunc),
	}
}

// Register 注册处理器
func (r *HandlerRegistry) Register(funcName string, handler HandlerFunc) {
	r.handlers[funcName] = handler
}

// Get 获取处理器
func (r *HandlerRegistry) Get(funcName string) (HandlerFunc, bool) {
	h, ok := r.handlers[funcName]
	return h, ok
}

// HandlerCount 返回已注册的处理器数量
func (r *HandlerRegistry) HandlerCount() int {
	return len(r.handlers)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// Start 启动 WebSocket 服务器
func Start(addr string, registry *HandlerRegistry, st *store.Store) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleConnection(w, r, registry, st)
	})

	return http.ListenAndServe(addr, nil)
}

// handleConnection 处理新的 WebSocket 连接
func handleConnection(w http.ResponseWriter, r *http.Request, registry *HandlerRegistry, st *store.Store) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket 升级失败: %v", err)
		return
	}

	session := &Session{
		Conn:    conn,
		Store:   st,
		Handler: registry,
		sendCh:  make(chan interface{}, 256),
		closeCh: make(chan struct{}),
	}

	log.Printf("🔗 客户端已连接: %s", conn.RemoteAddr())

	// 启动读写 goroutine
	go session.writePump()
	session.readPump()

	log.Printf("🔌 客户端已断开: %s", conn.RemoteAddr())
}

// SendMessage 线程安全地向客户端发送消息
func (s *Session) SendMessage(msg interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	select {
	case s.sendCh <- msg:
		return nil
	default:
		return nil
	}
}

// writePump 从发送队列取消息写入连接
func (s *Session) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()
	}()

	for {
		select {
		case msg, ok := <-s.sendCh:
			if !ok {
				return
			}
			s.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.Conn.WriteJSON(msg); err != nil {
				log.Printf("⚠️ 写入消息失败: %v", err)
				return
			}

		case <-ticker.C:
			s.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-s.closeCh:
			return
		}
	}
}

// readPump 读取客户端消息并分发
func (s *Session) readPump() {
	defer func() {
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()
		close(s.closeCh)
		s.Conn.Close()
	}()

	s.Conn.SetReadLimit(64 * 1024 * 1024) // 64MB 限制

	for {
		_, data, err := s.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("⚠️ 读取消息异常: %v", err)
			}
			return
		}

		// 解析请求
		funcName, iden, raw, err := protocol.ParseRequest(data)
		if err != nil {
			log.Printf("⚠️ 解析消息失败: %v", err)
			protocol.SendError(s.Conn, "", "", -1, err.Error())
			continue
		}

		if s.Store.Cfg.Verbose {
			log.Printf("📩 收到请求: func=%s iden=%s", funcName, iden)
		}

		// 查找并调用处理器
		handler, ok := s.Handler.Get(funcName)
		if !ok {
			log.Printf("⚠️ 未知的 func: %s", funcName)
			protocol.SendError(s.Conn, funcName, iden, -1, "unknown func: "+funcName)
			continue
		}

		// 在 goroutine 中处理，避免阻塞读取
		go handler(s, iden, raw)
	}
}

// ─── 辅助方法 (供 handler 使用) ───

// SendResponse 发送响应消息
func (s *Session) SendResponse(funcName, iden string, ret int, errInfo string, extra map[string]interface{}) {
	msg := protocol.Message{
		"func": funcName,
		"iden": iden,
		"ret":  ret,
	}
	if errInfo != "" {
		msg["err_info"] = errInfo
	}
	for k, v := range extra {
		msg[k] = v
	}
	s.SendMessage(msg)
}

// SendOK 发送成功响应
func (s *Session) SendOK(funcName, iden string, extra map[string]interface{}) {
	s.SendResponse(funcName, iden, 0, "", extra)
}

// SendError 发送错误响应
func (s *Session) SendError(funcName, iden string, ret int, errInfo string) {
	s.SendResponse(funcName, iden, ret, errInfo, nil)
}

// SendEvent 发送事件消息
func (s *Session) SendEvent(funcName, iden string, extra map[string]interface{}) {
	msg := protocol.Message{
		"func": funcName,
		"iden": iden,
	}
	for k, v := range extra {
		msg[k] = v
	}
	s.SendMessage(msg)
}
