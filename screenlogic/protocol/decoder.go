package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type Decoder struct {
	buffer *bytes.Buffer
}

func NewDecoder(buf *bytes.Buffer) *Decoder {
	return &Decoder{
		buffer: buf,
	}
}

func (d *Decoder) ReadBool() (bool, error) {
	b, err := d.buffer.ReadByte()
	if err != nil {
		return false, err
	}

	if b == 1 {
		return true, nil
	}

	return false, nil
}

func (d *Decoder) ReadUint8() (uint8, error) {
	return d.buffer.ReadByte()
}

func (d *Decoder) ReadUint16() (uint16, error) {
	var val uint16

	return val, binary.Read(d.buffer, binary.LittleEndian, &val)
}

func (d *Decoder) ReadUint32() (uint32, error) {
	var val uint32

	return val, binary.Read(d.buffer, binary.LittleEndian, &val)
}

var TruncatedPacketError = errors.New("truncated packet")

func (d *Decoder) Read(data []byte) (int, error) {
	if d.buffer.Len() < len(data) {
		return 0, TruncatedPacketError
	}

	return d.buffer.Read(data)
}

func (d *Decoder) ReadString() (string, error) {
	var len uint32

	err := binary.Read(d.buffer, binary.LittleEndian, &len)
	if err != nil {
		return "", err
	}

	if d.buffer.Len() < int(len) {
		return "", TruncatedPacketError
	}

	// Packet strings are encoded as length then value, where value is padded in
	// multiples of 4 bytes. Make sure we read the padded bytes as well in case
	// there is more data after in this packet. This will ensure we've skipped to
	// the correct place in the buffer for the next read.
	pad := 4 - (len % 4)

	if pad == 4 {
		pad = 0
	}

	buf := make([]byte, len+pad)

	n, err := d.buffer.Read(buf)
	if err != nil {
		return "", err
	}

	if n < int(len) {
		return "", TruncatedPacketError
	}

	// Only return the part of the buffer we care about, leaving the padded bytes off the end.
	return string(buf[:len]), nil
}

func (d *Decoder) CopyBytes(data []byte) error {
	n, err := d.buffer.Read(data)
	if err != nil {
		return err
	}

	if n < len(data) {
		return TruncatedPacketError
	}

	return nil
}

func (d *Decoder) ReadTail() ([]byte, error) {
	return d.buffer.Bytes(), nil
}
