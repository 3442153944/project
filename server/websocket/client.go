package websocket

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sunyuanling/server/pkg/logger"
	"go.uber.org/zap"
)

// Client WebSocket客户端
type Client struct {
	URL           string
	UserID        uint
	Token         string
	Device        *DeviceInfo
	conn          *websocket.Conn
	sendChan      chan []byte
	receiveChan   chan []byte
	isConnected   bool
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	reconnectWait time.Duration
	maxReconnect  int
	onConnect     func()
	onDisconnect  func()
	onMessage     func(*Message)
	metadata      map[string]interface{}
}

// ClientConfig 客户端配置
type ClientConfig struct {
	URL           string
	UserID        uint
	Token         string
	Device        *DeviceInfo
	ReconnectWait time.Duration
	MaxReconnect  int
}

// NewClient 创建客户端
func NewClient(config *ClientConfig) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	if config.ReconnectWait == 0 {
		config.ReconnectWait = 5 * time.Second
	}
	if config.MaxReconnect == 0 {
		config.MaxReconnect = 10
	}

	if config.Device == nil {
		config.Device = &DeviceInfo{
			DeviceID:   generateUUID(),
			DeviceType: DeviceTypeUnknown,
			Status:     DeviceStatusOnline,
		}
	}

	return &Client{
		URL:           config.URL,
		UserID:        config.UserID,
		Token:         config.Token,
		Device:        config.Device,
		sendChan:      make(chan []byte, 256),
		receiveChan:   make(chan []byte, 256),
		ctx:           ctx,
		cancel:        cancel,
		reconnectWait: config.ReconnectWait,
		maxReconnect:  config.MaxReconnect,
		metadata:      make(map[string]interface{}),
	}
}

// Connect 连接到服务器
func (c *Client) Connect() error {
	return c.connectWithRetry()
}

// connectWithRetry 带重试的连接
func (c *Client) connectWithRetry() error {
	retryCount := 0
	for {
		err := c.doConnect()
		if err == nil {
			retryCount = 0
			c.startReadWrite()
			return nil
		}

		if retryCount >= c.maxReconnect {
			logger.Error("达到最大重连次数",
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

		logger.Warn("连接失败，准备重试",
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
	// 构建URL参数
	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("device_id", c.Device.DeviceID)
	q.Set("device_type", string(c.Device.DeviceType))
	q.Set("device_name", c.Device.DeviceName)
	q.Set("platform", c.Device.Platform)
	q.Set("app_version", c.Device.AppVersion)
	u.RawQuery = q.Encode()

	// 设置请求头
	header := make(map[string][]string)
	header["Token"] = []string{c.Token}
	header["User-ID"] = []string{strconv.Itoa(int(c.UserID))}

	// 建立连接
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(u.String(), header)
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
		zap.String("device_id", c.Device.DeviceID),
	)

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
	defer c.handleDisconnect()

	c.conn.SetReadLimit(MaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(PongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, data, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Error("WebSocket读取错误",
						zap.Uint("user_id", c.UserID),
						zap.Error(err),
					)
				}
				return
			}

			// 解析消息
			if c.onMessage != nil {
				var msg Message
				if err := json.Unmarshal(data, &msg); err == nil {
					go c.onMessage(&msg)
				}
			}

			// 放入接收通道
			select {
			case c.receiveChan <- data:
			default:
				logger.Warn("接收通道已满",
					zap.Uint("user_id", c.UserID),
				)
			}
		}
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.handleDisconnect()
	}()

	for {
		select {
		case message, ok := <-c.sendChan:
			_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
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
			_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// heartbeatPump 发送心跳
func (c *Client) heartbeatPump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			msg := NewMessage(MessageTypeHeartbeat, nil)
			msg.From = &Sender{
				UserID:   c.UserID,
				DeviceID: c.Device.DeviceID,
			}
			_ = c.SendMessage(msg)

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

	if c.conn != nil {
		_ = c.conn.Close()
	}

	logger.Warn("WebSocket客户端断开连接",
		zap.Uint("user_id", c.UserID),
		zap.String("device_id", c.Device.DeviceID),
	)

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
			_ = c.connectWithRetry()
		}
	}()
}

// Send 发送原始数据
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

// SendMessage 发送消息
func (c *Client) SendMessage(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.Send(data)
}

// SendToUser 发送消息给用户
func (c *Client) SendToUser(userID uint, msgType MessageType, content interface{}) error {
	msg := NewMessage(msgType, content)
	msg.From = &Sender{
		UserID:   c.UserID,
		DeviceID: c.Device.DeviceID,
	}
	msg.Target = NewTargetUser(userID)
	return c.SendMessage(msg)
}

// SendToDevice 发送消息给设备
func (c *Client) SendToDevice(deviceID string, msgType MessageType, content interface{}) error {
	msg := NewMessage(msgType, content)
	msg.From = &Sender{
		UserID:   c.UserID,
		DeviceID: c.Device.DeviceID,
	}
	msg.Target = NewTargetDevice(deviceID)
	return c.SendMessage(msg)
}

// Receive 接收消息通道
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
func (c *Client) SetEventHandlers(onConnect, onDisconnect func(), onMessage func(*Message)) {
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

// GetDeviceInfo 获取设备信息
func (c *Client) GetDeviceInfo() *DeviceInfo {
	return c.Device
}

// SetDeviceStatus 设置设备状态
func (c *Client) SetDeviceStatus(status DeviceStatus) {
	c.Device.Status = status
}

// Close 关闭客户端
func (c *Client) Close() {
	c.cancel()

	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.isConnected = false
	c.mu.Unlock()

	close(c.sendChan)
	close(c.receiveChan)

	logger.Info("WebSocket客户端已关闭",
		zap.Uint("user_id", c.UserID),
		zap.String("device_id", c.Device.DeviceID),
	)
}

// Reconnect 手动重连
func (c *Client) Reconnect() {
	c.mu.Lock()
	if c.isConnected {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	go func() {
		_ = c.connectWithRetry()
	}()
}
