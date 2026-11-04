package voice

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type UDPConnection struct {
	conn *net.UDPConn
	SSRC uint32
	IP   string
	Port int
}

func NewUDPConnection(ip string, port int, ssrc uint32) (*UDPConnection, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	return &UDPConnection{
		conn: conn,
		SSRC: ssrc,
		IP:   ip,
		Port: port,
	}, nil
}

func (u *UDPConnection) Write(data []byte) error {
	_, err := u.conn.Write(data)
	return err
}

func (u *UDPConnection) Read() ([]byte, error) {
	buf := make([]byte, 2048)
	// Set a 30-second deadline so we can detect stuck reads
	u.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := u.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (u *UDPConnection) Close() error {
	return u.conn.Close()
}

func (u *UDPConnection) DiscoverIP() (string, int, error) {
	return u.DiscoverIPContext(context.Background())
}

// DiscoverIPContext performs UDP address discovery with cancellation and a
// bounded read deadline so an unresponsive voice endpoint cannot block forever.
func (u *UDPConnection) DiscoverIPContext(ctx context.Context) (string, int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	packet := make([]byte, 74)
	binary.BigEndian.PutUint16(packet[0:2], 1)
	binary.BigEndian.PutUint16(packet[2:4], 70)
	binary.BigEndian.PutUint32(packet[4:8], u.SSRC)

	if err := u.Write(packet); err != nil {
		return "", 0, err
	}

	buf := make([]byte, 74)
	deadline := time.Now().Add(5 * time.Second)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	if err := u.conn.SetReadDeadline(deadline); err != nil {
		return "", 0, err
	}
	stopCancel := context.AfterFunc(ctx, func() {
		_ = u.conn.SetReadDeadline(time.Now())
	})
	defer stopCancel()
	if _, err := u.conn.Read(buf); err != nil {
		if ctx.Err() != nil {
			return "", 0, ctx.Err()
		}
		if deadline, ok := ctx.Deadline(); ok && !time.Now().Before(deadline) {
			return "", 0, context.DeadlineExceeded
		}
		return "", 0, err
	}

	ipBytes := buf[8:72]
	ipLen := 0
	for i, b := range ipBytes {
		if b == 0 {
			ipLen = i
			break
		}
	}
	ip := string(ipBytes[:ipLen])

	port := int(binary.BigEndian.Uint16(buf[72:74]))

	return ip, port, nil
}
