package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/chainspace/simulations/pkg/engine"
	"github.com/chainspace/simulations/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// WebSocketHub WebSocket连接管理
type WebSocketHub struct {
	mu         sync.RWMutex
	clients    map[*WebSocketClient]bool
	broadcast  chan []byte
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	engine     *engine.Engine
	stop       chan struct{}
}

// WebSocketClient WebSocket客户端
type WebSocketClient struct {
	hub     *WebSocketHub
	conn    *websocket.Conn
	send    chan []byte
	filters map[string]bool // 事件过滤器
}

// NewWebSocketHub 创建WebSocket Hub
func NewWebSocketHub(eng *engine.Engine) *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		engine:     eng,
		stop:       make(chan struct{}),
	}
}

// Run 运行Hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case <-h.stop:
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Stop 停止Hub
func (h *WebSocketHub) Stop() {
	close(h.stop)
	h.mu.Lock()
	for client := range h.clients {
		close(client.send)
	}
	h.mu.Unlock()
}

// BroadcastEvent 广播事件
func (h *WebSocketHub) BroadcastEvent(event types.Event) {
	msg := WSMessage{
		Type: "event",
		Data: event,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case h.broadcast <- data:
	default:
	}
}

// BroadcastState 广播状态更新
func (h *WebSocketHub) BroadcastState(state *types.State) {
	msg := WSMessage{
		Type: "state_update",
		Data: state,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case h.broadcast <- data:
	default:
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WebSocketHub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &WebSocketClient{
		hub:     h,
		conn:    conn,
		send:    make(chan []byte, 256),
		filters: make(map[string]bool),
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

// WSMessage WebSocket消息格式
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// WSCommand WebSocket命令
type WSCommand struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// readPump 读取消息
func (c *WebSocketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var cmd WSCommand
		if err := json.Unmarshal(message, &cmd); err != nil {
			continue
		}

		c.handleCommand(cmd)
	}
}

// writePump 写入消息
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送缓冲区中的消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleCommand 处理命令
func (c *WebSocketClient) handleCommand(cmd WSCommand) {
	var response WSMessage

	switch cmd.Action {
	case "step":
		state, err := c.hub.engine.Step()
		if err != nil {
			response = WSMessage{Type: "error", Data: err.Error()}
		} else {
			response = WSMessage{Type: "state_update", Data: state}
		}

	case "pause":
		if err := c.hub.engine.Pause(); err != nil {
			response = WSMessage{Type: "error", Data: err.Error()}
		} else {
			response = WSMessage{Type: "status", Data: "paused"}
		}

	case "resume":
		if err := c.hub.engine.Resume(); err != nil {
			response = WSMessage{Type: "error", Data: err.Error()}
		} else {
			response = WSMessage{Type: "status", Data: "resumed"}
		}

	case "set_param":
		key, _ := cmd.Params["key"].(string)
		value := cmd.Params["value"]
		if err := c.hub.engine.SetParam(key, value); err != nil {
			response = WSMessage{Type: "error", Data: err.Error()}
		} else {
			response = WSMessage{Type: "param_set", Data: map[string]interface{}{"key": key, "value": value}}
		}

	case "set_speed":
		speed, _ := cmd.Params["speed"].(float64)
		if err := c.hub.engine.SetSpeed(speed); err != nil {
			response = WSMessage{Type: "error", Data: err.Error()}
		} else {
			response = WSMessage{Type: "speed_set", Data: speed}
		}

	case "inject_fault":
		faultType, _ := cmd.Params["type"].(string)
		target, _ := cmd.Params["target"].(string)
		fault := &types.Fault{
			Type:   types.FaultType(faultType),
			Target: types.NodeID(target),
			Params: cmd.Params,
			Active: true,
		}
		if err := c.hub.engine.InjectFault(fault); err != nil {
			response = WSMessage{Type: "error", Data: err.Error()}
		} else {
			response = WSMessage{Type: "fault_injected", Data: fault.ID}
		}

	case "inject_attack":
		attackType, _ := cmd.Params["type"].(string)
		target, _ := cmd.Params["target"].(string)
		attack := &types.Attack{
			Type:   types.AttackType(attackType),
			Target: target,
			Params: cmd.Params,
			Active: true,
		}
		if err := c.hub.engine.InjectAttack(attack); err != nil {
			response = WSMessage{Type: "error", Data: err.Error()}
		} else {
			response = WSMessage{Type: "attack_injected", Data: attack.ID}
		}

	case "get_state":
		state := c.hub.engine.GetState()
		response = WSMessage{Type: "state_update", Data: state}

	case "subscribe":
		eventType, _ := cmd.Params["event_type"].(string)
		c.filters[eventType] = true
		response = WSMessage{Type: "subscribed", Data: eventType}

	case "unsubscribe":
		eventType, _ := cmd.Params["event_type"].(string)
		delete(c.filters, eventType)
		response = WSMessage{Type: "unsubscribed", Data: eventType}

	default:
		response = WSMessage{Type: "error", Data: "unknown action"}
	}

	data, _ := json.Marshal(response)
	c.send <- data
}

// ClientCount 获取客户端数量
func (h *WebSocketHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
