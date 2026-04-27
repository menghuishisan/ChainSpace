package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeEnvStatus     MessageType = "env_status"
	MessageTypeSessionUpdate MessageType = "session_update"
	MessageTypeSessionChat   MessageType = "session_chat"
	MessageTypeEnvOutput     MessageType = "env_output"
	MessageTypeNotification  MessageType = "notification"
	MessageTypeContestUpdate MessageType = "contest_update"
	MessageTypeScoreboard    MessageType = "scoreboard"
	MessageTypePing          MessageType = "ping"
	MessageTypePong          MessageType = "pong"
)

// Message WebSocket消息
type Message struct {
	Type      MessageType     `json:"type"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// Client WebSocket客户端
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   uint
	schoolID uint
	rooms    map[string]bool
	mu       sync.RWMutex
}

// Hub WebSocket中心
type Hub struct {
	clients    map[*Client]bool
	userIndex  map[uint]*Client
	roomIndex  map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	Room    string
	Message []byte
}

// NewHub 创建Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		userIndex:  make(map[uint]*Client),
		roomIndex:  make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

// Run 运行Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.userIndex[client.userID] = client
			h.mu.Unlock()
			logger.Debug("Client registered", zap.Uint("userID", client.userID))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.mu.Lock()
				delete(h.clients, client)
				delete(h.userIndex, client.userID)
				for room := range client.rooms {
					if roomClients, ok := h.roomIndex[room]; ok {
						delete(roomClients, client)
						if len(roomClients) == 0 {
							delete(h.roomIndex, room)
						}
					}
				}
				h.mu.Unlock()
				close(client.send)
				logger.Debug("Client unregistered", zap.Uint("userID", client.userID))
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			if roomClients, ok := h.roomIndex[message.Room]; ok {
				for client := range roomClients {
					select {
					case client.send <- message.Message:
					default:
						h.mu.RUnlock()
						h.unregister <- client
						h.mu.RLock()
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterClient 注册客户端
func (h *Hub) RegisterClient(conn *websocket.Conn, userID, schoolID uint) *Client {
	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		schoolID: schoolID,
		rooms:    make(map[string]bool),
	}

	h.register <- client
	return client
}

// JoinRoom 加入房间
func (h *Hub) JoinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	client.rooms[room] = true
	client.mu.Unlock()

	if _, ok := h.roomIndex[room]; !ok {
		h.roomIndex[room] = make(map[*Client]bool)
	}
	h.roomIndex[room][client] = true

	logger.Debug("Client joined room", zap.Uint("userID", client.userID), zap.String("room", room))
}

// LeaveRoom 离开房间
func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	delete(client.rooms, room)
	client.mu.Unlock()

	if roomClients, ok := h.roomIndex[room]; ok {
		delete(roomClients, client)
		if len(roomClients) == 0 {
			delete(h.roomIndex, room)
		}
	}
}

// BroadcastToRoom 广播到房间
func (h *Hub) BroadcastToRoom(room string, msgType MessageType, data interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	msg := Message{
		Type:      msgType,
		Data:      dataJSON,
		Timestamp: time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	h.broadcast <- &BroadcastMessage{
		Room:    room,
		Message: msgBytes,
	}

	return nil
}

// SendToUser 发送给指定用户
func (h *Hub) SendToUser(userID uint, msgType MessageType, data interface{}) error {
	h.mu.RLock()
	client, ok := h.userIndex[userID]
	h.mu.RUnlock()

	if !ok {
		return nil // 用户不在线
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	msg := Message{
		Type:      msgType,
		Data:      dataJSON,
		Timestamp: time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case client.send <- msgBytes:
	default:
		h.unregister <- client
	}

	return nil
}

// GetOnlineCount 获取在线人数
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetRoomCount 获取房间在线人数
func (h *Hub) GetRoomCount(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if roomClients, ok := h.roomIndex[room]; ok {
		return len(roomClients)
	}
	return 0
}

// GetClientByUserID 根据用户ID获取客户端
func (h *Hub) GetClientByUserID(userID uint) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.userIndex[userID]
}

// ReadPump 读取消息泵
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket read error", zap.Error(err))
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		c.handleMessage(&msg)
	}
}

// WritePump 写入消息泵
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理消息
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case MessageTypePing:
		c.sendPong()
	default:
		// 处理其他消息类型
	}
}

// sendPong 发送pong
func (c *Client) sendPong() {
	msg := Message{
		Type:      MessageTypePong,
		Timestamp: time.Now().Unix(),
	}
	msgBytes, _ := json.Marshal(msg)
	c.send <- msgBytes
}
