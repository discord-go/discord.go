package voice

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/discord-go/discord.go/snowflake"
)

type mockConn struct {
	readData [][]byte
	readErrs []error
	readIdx  int

	writeData [][]byte
	writeErr  error
	closeErr  error
	closed    bool
	blockRead chan struct{}
}

func (m *mockConn) Read() ([]byte, error) {
	if m.readIdx >= len(m.readData) {
		if m.blockRead != nil {
			<-m.blockRead
		}
		if m.readIdx < len(m.readData)+len(m.readErrs) {
			err := m.readErrs[m.readIdx-len(m.readData)]
			m.readIdx++
			return nil, err
		}
		return nil, errors.New("EOF")
	}
	data := m.readData[m.readIdx]
	m.readIdx++
	return data, nil
}

func (m *mockConn) Write(data []byte) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.writeData = append(m.writeData, data)
	return nil
}

func (m *mockConn) Close() error {
	m.closed = true
	return m.closeErr
}

func (m *mockConn) ReadMessage() (int, []byte, error) {
	data, err := m.Read()
	return MsgTypeText, data, err
}

func (m *mockConn) WriteBinary(data []byte) error {
	return m.Write(data)
}

func startLocalUDPServer(t *testing.T) (string, int, func()) {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	serverPort := serverConn.LocalAddr().(*net.UDPAddr).Port
	done := make(chan struct{})

	go func() {
		buf := make([]byte, 2048)
		for {
			select {
			case <-done:
				return
			default:
			}
			serverConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, addr, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			if n == 74 {
				resp := make([]byte, 74)
				copy(resp[8:], []byte("127.0.0.1"))
				resp[72] = byte(12345 >> 8)
				resp[73] = byte(12345 & 0xFF)
				serverConn.WriteToUDP(resp, addr)
			}
		}
	}()

	return "127.0.0.1", serverPort, func() {
		close(done)
		serverConn.Close()
	}
}

func buildPayload(op int, data interface{}) []byte {
	b, _ := json.Marshal(data)
	p, _ := json.Marshal(Payload{Op: op, Data: b})
	return p
}

func TestNewClient(t *testing.T) {
	conn := &mockConn{}
	guildID := snowflake.ID(1)
	channelID := snowflake.ID(2)
	userID := snowflake.ID(3)

	c := NewClient(conn, guildID, channelID, userID)

	if c.Conn != conn {
		t.Error("expected Conn to be set")
	}
	if c.GuildID != guildID {
		t.Errorf("expected GuildID %v, got %v", guildID, c.GuildID)
	}
	if c.ChannelID != channelID {
		t.Errorf("expected ChannelID %v, got %v", channelID, c.ChannelID)
	}
	if c.UserID != userID {
		t.Errorf("expected UserID %v, got %v", userID, c.UserID)
	}
	if c.State != StateIdle {
		t.Errorf("expected State %v, got %v", StateIdle, c.State)
	}
}

func TestClient_Connect(t *testing.T) {
	ip, port, cleanup := startLocalUDPServer(t)
	defer cleanup()

	conn := &mockConn{
		readData: [][]byte{
			buildPayload(Hello, HelloData{HeartbeatInterval: 1000}),
			buildPayload(Ready, ReadyData{IP: ip, Port: port, SSRC: 1234}),
			buildPayload(SessionDescription, SessionDescriptionData{SecretKey: make([]byte, 32)}),
		},
		blockRead: make(chan struct{}),
	}
	defer close(conn.blockRead)
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))
	c.Token = "test-bypass"

	err := c.Connect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.State != StateConnected {
		t.Errorf("expected State %v, got %v", StateConnected, c.State)
	}
	if len(conn.writeData) < 2 {
		t.Fatalf("expected at least 2 writes, got %d", len(conn.writeData))
	}

	var p Payload
	json.Unmarshal(conn.writeData[0], &p)
	if p.Op != Identify {
		t.Errorf("expected Identify, got %d", p.Op)
	}
	json.Unmarshal(conn.writeData[1], &p)
	if p.Op != SelectProtocol {
		t.Errorf("expected SelectProtocol, got %d", p.Op)
	}
}

func TestClient_Connect_NotIdle(t *testing.T) {
	conn := &mockConn{}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))
	c.State = StateConnected

	err := c.Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for non-idle state")
	}
	if err.Error() != "voice: client is not idle" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_Connect_ContextCancelled(t *testing.T) {
	conn := &mockConn{}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.Connect(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if c.State != StateIdle {
		t.Errorf("expected State %v after cancelled connect, got %v", StateIdle, c.State)
	}
}

func TestClient_Connect_WriteError(t *testing.T) {
	conn := &mockConn{
		readData: [][]byte{
			buildPayload(Hello, HelloData{HeartbeatInterval: 1000}),
		},
		writeErr: errors.New("write failed"),
	}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))

	err := c.Connect(context.Background())
	if err == nil {
		t.Fatal("expected write error")
	}
	if err.Error() != "write failed" {
		t.Errorf("unexpected error: %v", err)
	}
	if c.State != StateIdle {
		t.Errorf("expected State %v after write error, got %v", StateIdle, c.State)
	}
}

func TestClient_Disconnect(t *testing.T) {
	conn := &mockConn{}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))
	c.State = StateConnected
	c.cancelHB = func() {}
	c.UDP, _ = NewUDPConnection("127.0.0.1", 12345, 1234)

	err := c.Disconnect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.State != StateIdle {
		t.Errorf("expected State %v, got %v", StateIdle, c.State)
	}
	if !conn.closed {
		t.Error("expected connection to be closed")
	}
}

func TestClient_Disconnect_AlreadyIdle(t *testing.T) {
	conn := &mockConn{}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))

	err := c.Disconnect()
	if err == nil {
		t.Fatal("expected error for idle disconnect")
	}
	if err.Error() != "voice: client is already idle" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_Disconnect_CloseError(t *testing.T) {
	conn := &mockConn{closeErr: errors.New("close failed")}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))
	c.State = StateConnected
	c.cancelHB = func() {}

	err := c.Disconnect()
	if err == nil {
		t.Fatal("expected close error")
	}
	if err.Error() != "close failed" {
		t.Errorf("unexpected error: %v", err)
	}
	if c.State != StateIdle {
		t.Errorf("expected State %v, got %v", StateIdle, c.State)
	}
}

func TestClient_Reconnect_NotConnected(t *testing.T) {
	conn := &mockConn{}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))

	err := c.Reconnect(context.Background())
	if err == nil {
		t.Fatal("expected error for non-connected reconnect")
	}
	if err.Error() != "voice: client is not connected" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_Reconnect_CloseError(t *testing.T) {
	conn := &mockConn{closeErr: errors.New("close error")}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))
	c.State = StateConnected
	c.cancelHB = func() {}

	err := c.Reconnect(context.Background())
	if err == nil {
		t.Fatal("expected close error")
	}
	if err.Error() != "close error" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_SetSession(t *testing.T) {
	conn := &mockConn{}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))

	c.SetSession("session-123", "token-abc", "us-west.discord.gg")

	if c.SessionID != "session-123" {
		t.Errorf("expected SessionID %q, got %q", "session-123", c.SessionID)
	}
	if c.Token != "token-abc" {
		t.Errorf("expected Token %q, got %q", "token-abc", c.Token)
	}
	if c.Endpoint != "us-west.discord.gg" {
		t.Errorf("expected Endpoint %q, got %q", "us-west.discord.gg", c.Endpoint)
	}
}

func TestClient_GetState(t *testing.T) {
	conn := &mockConn{}
	c := NewClient(conn, snowflake.ID(1), snowflake.ID(2), snowflake.ID(3))

	if got := c.GetState(); got != StateIdle {
		t.Errorf("expected State %v, got %v", StateIdle, got)
	}

	c.State = StateConnected
	if got := c.GetState(); got != StateConnected {
		t.Errorf("expected State %v, got %v", StateConnected, got)
	}
}
