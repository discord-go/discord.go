package voice

import (
	"encoding/binary"
	"errors"
)

type RTPHeader struct {
	Version     uint8
	Padding     bool
	Extension   bool
	CSRCCount   uint8
	Marker      bool
	PayloadType uint8
	Sequence    uint16
	Timestamp   uint32
	SSRC        uint32
}

func NewRTPHeader(sequence uint16, timestamp uint32, ssrc uint32) *RTPHeader {
	return &RTPHeader{
		Version:     2,
		PayloadType: 120, // Discord usually uses Opus payload type 120
		Sequence:    sequence,
		Timestamp:   timestamp,
		SSRC:        ssrc,
	}
}

func (h *RTPHeader) Marshal() []byte {
	buf := make([]byte, 12)
	buf[0] = (h.Version << 6) | (b2u8(h.Padding) << 5) | (b2u8(h.Extension) << 4) | (h.CSRCCount & 0x0f)
	buf[1] = (b2u8(h.Marker) << 7) | (h.PayloadType & 0x7f)
	binary.BigEndian.PutUint16(buf[2:4], h.Sequence)
	binary.BigEndian.PutUint32(buf[4:8], h.Timestamp)
	binary.BigEndian.PutUint32(buf[8:12], h.SSRC)
	return buf
}

func ParseRTPHeader(data []byte) (*RTPHeader, error) {
	if len(data) < 12 {
		return nil, errors.New("voice: packet too short")
	}
	version := data[0] >> 6
	if version != 2 {
		return nil, errors.New("voice: unsupported RTP version")
	}
	cc := int(data[0] & 0x0f)
	headerLength := 12 + 4*cc
	if len(data) < headerLength {
		return nil, errors.New("voice: packet shorter than CSRC header")
	}
	if data[0]&0x10 != 0 {
		if len(data) < headerLength+4 {
			return nil, errors.New("voice: packet missing RTP extension header")
		}
		extensionLength := int(binary.BigEndian.Uint16(data[headerLength+2:headerLength+4])) * 4
		headerLength += 4 + extensionLength
		if len(data) < headerLength {
			return nil, errors.New("voice: packet shorter than RTP extension")
		}
	}
	if data[0]&0x20 != 0 {
		padding := int(data[len(data)-1])
		if padding == 0 || padding > len(data)-headerLength {
			return nil, errors.New("voice: invalid RTP padding")
		}
	}

	return &RTPHeader{
		Version:     version,
		Padding:     (data[0] >> 5 & 1) == 1,
		Extension:   (data[0] >> 4 & 1) == 1,
		CSRCCount:   data[0] & 0x0f,
		Marker:      (data[1] >> 7 & 1) == 1,
		PayloadType: data[1] & 0x7f,
		Sequence:    binary.BigEndian.Uint16(data[2:4]),
		Timestamp:   binary.BigEndian.Uint32(data[4:8]),
		SSRC:        binary.BigEndian.Uint32(data[8:12]),
	}, nil
}

func BuildAudioPacket(header *RTPHeader, payload []byte) []byte {
	hdrBytes := header.Marshal()
	packet := make([]byte, len(hdrBytes)+len(payload))
	copy(packet, hdrBytes)
	copy(packet[len(hdrBytes):], payload)
	return packet
}

func b2u8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
