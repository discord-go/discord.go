package voice

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

func TestUDPConnection(t *testing.T) {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer serverConn.Close()

	serverPort := serverConn.LocalAddr().(*net.UDPAddr).Port

	go func() {
		buf := make([]byte, 2048)
		for {
			n, addr, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			if n == 74 {
				resp := make([]byte, 74)
				copy(resp[8:], []byte("127.0.0.1"))
				binary.BigEndian.PutUint16(resp[72:74], uint16(12345))
				serverConn.WriteToUDP(resp, addr)
			} else {
				serverConn.WriteToUDP(buf[:n], addr)
			}
		}
	}()

	conn, err := NewUDPConnection("127.0.0.1", serverPort, 1234)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	ip, port, err := conn.DiscoverIP()
	if err != nil {
		t.Fatal(err)
	}
	if ip != "127.0.0.1" || port != 12345 {
		t.Errorf("Expected 127.0.0.1:12345, got %s:%d", ip, port)
	}

	msg := []byte("hello")
	if err := conn.Write(msg); err != nil {
		t.Fatal(err)
	}

	conn.conn.SetReadDeadline(time.Now().Add(time.Second))
	resp, err := conn.Read()
	if err != nil {
		t.Fatal(err)
	}
	if string(resp) != string(msg) {
		t.Errorf("Expected %s, got %s", string(msg), string(resp))
	}
}
