// websocket_client.go
// Socket.io WebSocket client for ABS real-time events

package abs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient connects to ABS Socket.io for real-time events
type WebSocketClient struct {
	baseURL       string
	token         string
	headerFile    string
	inlineHeaders []string
	headers       map[string]string
	conn          *websocket.Conn
	connected     bool
	mu            sync.RWMutex

	// Event handlers
	scanStartHandlers    []func(LibraryScan)
	scanCompleteHandlers []func(LibraryScanResults)

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// LibraryScan represents a scan_start event
type LibraryScan struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// LibraryScanResults represents a scan_complete event results
type LibraryScanResults struct {
	Added   int `json:"added"`
	Updated int `json:"updated"`
	Missing int `json:"missing"`
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(
	baseURL, token, headerFile string,
	inlineHeaders []string,
) *WebSocketClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketClient{
		baseURL:       baseURL,
		token:         token,
		headerFile:    headerFile,
		inlineHeaders: inlineHeaders,
		headers:       make(map[string]string),
		ctx:           ctx,
		cancel:        cancel,
		done:          make(chan struct{}),
	}
}

// Connect establishes WebSocket connection and authenticates
func (w *WebSocketClient) Connect() error {
	// Load headers from file if specified
	if w.headerFile != "" {
		if err := w.loadHeadersFromFile(w.headerFile); err != nil {
			return fmt.Errorf("loading headers: %w", err)
		}
	}

	// Parse inline headers
	for _, h := range w.inlineHeaders {
		parts := strings.SplitN(h, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			w.headers[key] = value
		}
	}

	// Convert http:// to ws://
	wsURL := w.baseURL
	if len(wsURL) > 4 && wsURL[:4] == "http" {
		wsURL = "ws" + wsURL[4:]
	}

	// Parse URL and add socket.io path
	u, err := url.Parse(wsURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = "/socket.io/"
	u.RawQuery = "EIO=4&transport=websocket"

	// Prepare headers for WebSocket handshake
	headers := make(http.Header)
	for key, value := range w.headers {
		headers.Set(key, value)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	w.conn = conn

	// Wait for engine.io open packet ("0" or "0{...}")
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to read open packet: %w", err)
	}
	if len(msg) == 0 || msg[0] != '0' {
		conn.Close()
		return fmt.Errorf("unexpected open packet: %s", string(msg))
	}

	// Socket.io v4: Send auth as "40{"token":"..."}"
	// Format: 4 = message packet, 0 = connect event
	authMsg := fmt.Sprintf("40{\"token\":\"%s\"}", w.token)
	if err := w.conn.WriteMessage(websocket.TextMessage, []byte(authMsg)); err != nil {
		conn.Close()
		return fmt.Errorf("auth failed: %w", err)
	}

	w.mu.Lock()
	w.connected = true
	w.mu.Unlock()

	// Start message handler
	go w.handleMessages()

	return nil
}

// loadHeadersFromFile loads headers from a file in KEY=VALUE format
func (w *WebSocketClient) loadHeadersFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading header file: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		w.headers[key] = value
	}

	return nil
}

// Close disconnects the WebSocket
func (w *WebSocketClient) Close() error {
	w.cancel()
	w.mu.Lock()
	w.connected = false
	w.mu.Unlock()

	if w.conn != nil {
		w.conn.Close()
	}

	<-w.done
	return nil
}

// IsConnected returns connection status
func (w *WebSocketClient) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connected
}

// OnScanStart registers a handler for scan start events
func (w *WebSocketClient) OnScanStart(handler func(LibraryScan)) {
	w.scanStartHandlers = append(w.scanStartHandlers, handler)
}

// OnScanComplete registers a handler for scan complete events
func (w *WebSocketClient) OnScanComplete(handler func(LibraryScanResults)) {
	w.scanCompleteHandlers = append(w.scanCompleteHandlers, handler)
}

// WaitForScanComplete blocks until scan completes for a specific library or timeout
func (w *WebSocketClient) WaitForScanComplete(
	libraryID string,
	timeout time.Duration,
) (*LibraryScanResults, error) {
	if !w.IsConnected() {
		return nil, fmt.Errorf("websocket not connected")
	}

	scanning := false
	resultChan := make(chan LibraryScanResults, 1)

	// Track when our library starts scanning
	startHandler := func(scan LibraryScan) {
		if scan.ID == libraryID {
			scanning = true
		}
	}
	w.OnScanStart(startHandler)

	// Capture completion for our library
	completeHandler := func(results LibraryScanResults) {
		if scanning {
			select {
			case resultChan <- results:
			default:
			}
		}
	}
	w.OnScanComplete(completeHandler)

	// Wait with timeout
	select {
	case results := <-resultChan:
		return &results, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for scan completion (library: %s)", libraryID)
	case <-w.ctx.Done():
		return nil, fmt.Errorf("websocket disconnected")
	}
}

func (w *WebSocketClient) handleMessages() {
	defer close(w.done)

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		_, message, err := w.conn.ReadMessage()
		if err != nil {
			w.mu.Lock()
			w.connected = false
			w.mu.Unlock()
			return
		}

		// Parse Socket.io message
		// Format: <packet type><event name><data>
		w.parseMessage(message)
	}
}

func (w *WebSocketClient) parseMessage(data []byte) {
	if len(data) < 2 {
		return
	}

	// Socket.io packet types:
	// 0 = open, 1 = close, 2 = ping, 3 = pong, 4 = message, 5 = upgrade, 6 = noop
	packetType := data[0]

	// We only care about message packets (4)
	if packetType != '4' {
		return
	}

	// Message packet format: 4<event name>["data"]
	// Common formats:
	// 42["event_name", data] - event with data
	// 40{} - connect
	// 41 - disconnect

	if len(data) < 3 || data[1] != '2' {
		return // Not an event packet (42 = message + event)
	}

	// Parse the event payload: ["event_name", {data}]
	var payload []json.RawMessage
	if err := json.Unmarshal(data[2:], &payload); err != nil {
		return
	}

	if len(payload) < 1 {
		return
	}

	// First element is event name
	var eventName string
	if err := json.Unmarshal(payload[0], &eventName); err != nil {
		return
	}

	switch eventName {
	case "scan_start":
		if len(payload) < 2 {
			return
		}
		var scan LibraryScan
		if err := json.Unmarshal(payload[1], &scan); err == nil {
			for _, h := range w.scanStartHandlers {
				go h(scan)
			}
		}
	case "scan_complete":
		if len(payload) < 2 {
			return
		}
		var results LibraryScanResults
		if err := json.Unmarshal(payload[1], &results); err == nil {
			for _, h := range w.scanCompleteHandlers {
				go h(results)
			}
		}
	}
}
