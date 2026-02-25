package mcp

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TransportMap 存储了当前活跃的 SSE session
var TransportMap = sync.Map{}

// SSEServerTransport 针对 Gin 实现了 mcp.Transport 接口
type SSEServerTransport struct {
	SendChan chan any // 保持为 any 类型以提供 json.Marshal 灵活性, 但 Write 接收 jsonrpc.Message
	recvChan chan jsonrpc.Message
	id       string
	closed   bool
	mutex    sync.Mutex
}

func NewSSEServerTransport() *SSEServerTransport {
	t := &SSEServerTransport{
		SendChan: make(chan any, 10),
		recvChan: make(chan jsonrpc.Message, 10),
		id:       uuid.New().String(),
	}
	TransportMap.Store(t.id, t)
	return t
}

// Connect 实现 mcp.Transport 接口
func (t *SSEServerTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	return t, nil
}

// Read 实现 mcp.Connection 接口
func (t *SSEServerTransport) Read(ctx context.Context) (jsonrpc.Message, error) {
	select {
	case msg, ok := <-t.recvChan:
		if !ok {
			return nil, io.EOF
		}
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Write 实现 mcp.Connection 接口
func (t *SSEServerTransport) Write(ctx context.Context, message jsonrpc.Message) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return fmt.Errorf("transport closed")
	}
	select {
	case t.SendChan <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close 实现 mcp.Connection 接口
func (t *SSEServerTransport) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if !t.closed {
		t.closed = true
		close(t.SendChan)
		close(t.recvChan)
		TransportMap.Delete(t.id)
	}
	return nil
}

// SessionID 实现 mcp.Connection 接口
func (t *SSEServerTransport) SessionID() string {
	return t.id
}

// HandleMessage 将被 POST 路由调用来注入客户端发来的消息
func (t *SSEServerTransport) HandleMessage(msg jsonrpc.Message) {
	t.mutex.Lock()
	if t.closed {
		t.mutex.Unlock()
		return
	}
	t.mutex.Unlock()

	select {
	case t.recvChan <- msg:
	default:
		// 阻塞式的后备写入策略
		t.recvChan <- msg
	}
}
