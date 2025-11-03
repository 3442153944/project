package websocket

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"project/pkg/logger"
)

// Client WebSocket客户端（支持自动重连）
type Client struct {
	URL           string                 // WebSocket服务器URL
	UserID        uint                   // 用户ID
	Token         string                 // 认证Token
	conn          *websocket.Conn        // WebSocket连接
	sendChan      chan []byte            // 发送通道
	receiveChan   chan []byte            // 接收通道
	isConnected   bool                   // 连接状态
	mu            sync.RWMutex           // 读写锁
	ctx           context.Context        // 上下文
	cancel        context.CancelFunc     // 取消函数
	reconnectWait time.Duration          // 重连等待时间
	maxReconnect  int                    // 最大重连次数
	onConnect     func()                 // 连接成功回调
	onDisconnect  func()                 // 断开连接回调
	onMessage     func([]byte)           // 消息回调
	metadata      map[string]interface{} // 元数据
}

// ClientConfig 客户端配置
type ClientConfig struct {
	URL             string        // WebSocket服务器URL
	UserID          uint          // 用户ID
	Token           string        // 认证Token
	ReconnectWait   time.Duration // 重连等待时间
	MaxReconnect    int           // 最大重连次数
	HeartbeatPeriod time.Duration // 心跳周期
}

// NewClient 创建新的WebSocket客户端
func NewClient(config *ClientConfig) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	if config.ReconnectWait == 0 {
		config.ReconnectWait = 5 * time.Second
	}
	if config.MaxReconnect == 0 {
		config.MaxReconnect = 10
	}

	return &Client{
		URL:           config.URL,
		UserID:        config.UserID,
		Token:         config.Token,
		sendChan:      make(chan []byte, 256),
		receiveChan:   make(chan []byte, 256),
		ctx:           ctx,
		cancel:        cancel,
		reconnectWait: config.ReconnectWait,
		maxReconnect:  config.MaxReconnect,
		metadata:      make(map[string]interface{}),
	}
}

// Connect 连接到WebSocket服务器
func (c *Client) Connect() error {
	return c.connectWithRetry()
}

// connectWithRetry 带重试的连接
func (c *Client) connectWithRetry() error {
	retryCount := 0
	for {
		err := c.doConnect()
		if err == nil {
			// 连接成功
			retryCount = 0
			c.startReadWrite()
			return nil
		}

		if retryCount >= c.maxReconnect {
			logger.Error("达到最大重连次数，放弃连接",
				zap.Uint("user_id", c.UserID),
				zap.Int("max_reconnect", c.maxReconnect),
			)
			return err
		}

		retryCount++
		wait := time.Duration(retryCount) * c.reconnectWait
		if wait > 60*time.Second {
			wait = 60 * time.Second
		}

		logger.Warn("WebSocket连接失败，准备重试",
			zap.Uint("user_id", c.UserID),
			zap.Int("retry_count", retryCount),
			zap.Duration("wait", wait),
			zap.Error(err),
		)

		select {
		case <-time.After(wait):
			continue
		case <-c.ctx.Done():
			return context.Canceled
		}
	}
}

// doConnect 执行连接
func (c *Client) doConnect() error {
	// 设置请求头
	header := make(map[string][]string)
	header["Token"] = []string{c.Token}
	header["User-ID"] = []string{strconv.Itoa(int(c.UserID))}
	println("请求头:", header)

	// 建立连接
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(c.URL, header)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.isConnected = true
	c.mu.Unlock()

	logger.Info("WebSocket客户端连接成功",
		zap.String("url", c.URL),
		zap.Uint("user_id", c.UserID),
	)

	// 触发连接成功回调
	if c.onConnect != nil {
		go c.onConnect()
	}

	return nil
}

// startReadWrite 启动读写协程
func (c *Client) startReadWrite() {
	go c.readPump()
	go c.writePump()
	go c.heartbeatPump()
}

// readPump 读取消息
func (c *Client) readPump() {
	defer func() {
		c.handleDisconnect()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Error("WebSocket读取错误",
						zap.Uint("user_id", c.UserID),
						zap.Error(err),
					)
				}
				return
			}

			// 处理消息
			if c.onMessage != nil {
				go c.onMessage(message)
			}

			// 放入接收通道
			select {
			case c.receiveChan <- message:
			default:
				logger.Warn("接收通道已满，丢弃消息",
					zap.Uint("user_id", c.UserID),
				)
			}
		}
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.handleDisconnect()
	}()

	for {
		select {
		case message, ok := <-c.sendChan:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Error("WebSocket写入错误",
					zap.Uint("user_id", c.UserID),
					zap.Error(err),
				)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// heartbeatPump 心跳发送
func (c *Client) heartbeatPump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			msg := Message{
				Type:      MessageTypeHeartbeat,
				From:      c.UserID,
				Timestamp: time.Now().Unix(),
			}
			c.SendMessage(&msg)

		case <-c.ctx.Done():
			return
		}
	}
}

// handleDisconnect 处理断开连接
func (c *Client) handleDisconnect() {
	c.mu.Lock()
	if !c.isConnected {
		c.mu.Unlock()
		return
	}
	c.isConnected = false
	c.mu.Unlock()

	// 关闭连接
	if c.conn != nil {
		c.conn.Close()
	}

	logger.Warn("WebSocket客户端断开连接",
		zap.Uint("user_id", c.UserID),
	)

	// 触发断开连接回调
	if c.onDisconnect != nil {
		go c.onDisconnect()
	}

	// 尝试重连
	go func() {
		select {
		case <-c.ctx.Done():
			return
		default:
			time.Sleep(c.reconnectWait)
			err := c.connectWithRetry()
			if err != nil {
				return
			}
		}
	}()
}

// Send 发送消息
func (c *Client) Send(data []byte) error {
	c.mu.RLock()
	if !c.isConnected {
		c.mu.RUnlock()
		return ErrConnectionClosed
	}
	c.mu.RUnlock()

	select {
	case c.sendChan <- data:
		return nil
	case <-time.After(5 * time.Second):
		return ErrSendTimeout
	}
}

// SendMessage 发送消息对象
func (c *Client) SendMessage(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.Send(data)
}

// SendJSON 发送JSON数据
func (c *Client) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Send(data)
}

// Receive 接收消息
func (c *Client) Receive() <-chan []byte {
	return c.receiveChan
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// SetEventHandlers 设置事件处理器
func (c *Client) SetEventHandlers(onConnect, onDisconnect func(), onMessage func([]byte)) {
	c.onConnect = onConnect
	c.onDisconnect = onDisconnect
	c.onMessage = onMessage
}

// SetMetadata 设置元数据
func (c *Client) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	c.metadata[key] = value
	c.mu.Unlock()
}

// GetMetadata 获取元数据
func (c *Client) GetMetadata(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.metadata[key]
	return value, exists
}

// Close 关闭客户端
func (c *Client) Close() {
	c.cancel()

	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.isConnected = false
	c.mu.Unlock()

	close(c.sendChan)
	close(c.receiveChan)

	logger.Info("WebSocket客户端已关闭",
		zap.Uint("user_id", c.UserID),
	)
}

// Reconnect 手动触发重连
func (c *Client) Reconnect() {
	c.mu.Lock()
	if c.isConnected {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	go c.connectWithRetry()
}
