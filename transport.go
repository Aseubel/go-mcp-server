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

// TransportMap holds active SSE sessions
var TransportMap = sync.Map{}

// SSEServerTransport implements mcp.Transport for Gin
type SSEServerTransport struct {
	SendChan chan any // Keeping as any for json.Marshal flexibility, but Write takes jsonrpc.Message
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

// Connect implements mcp.Transport
func (t *SSEServerTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	return t, nil
}

// Read implements mcp.Connection
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

// Write implements mcp.Connection
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

// Close implements mcp.Connection
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

// SessionID implements mcp.Connection
func (t *SSEServerTransport) SessionID() string {
	return t.id
}

// HandleMessage is called by the POST handler to inject messages
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
		// Blocking fallback
		t.recvChan <- msg
	}
}
