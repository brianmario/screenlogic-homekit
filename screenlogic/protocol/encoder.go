package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
)

type Encoder struct {
	buffer *bytes.Buffer
}

func NewEncoder(buf *bytes.Buffer) *Encoder {
	return &Encoder{
		buffer: buf,
	}
}

func (e *Encoder) WriteUint8(val uint8) error {
	return binary.Write(e.buffer, binary.LittleEndian, val)
}

func (e *Encoder) WriteUint16(val uint16) error {
	return binary.Write(e.buffer, binary.LittleEndian, val)
}

func (e *Encoder) WriteUint32(val uint32) error {
	return binary.Write(e.buffer, binary.LittleEndian, val)
}

func (e *Encoder) WriteString(val string) error {
	len := uint32(len(val))

	err := binary.Write(e.buffer, binary.LittleEndian, len)
	if err != nil {
		return err
	}

	n, err := e.buffer.WriteString(val)
	if err != nil {
		return err
	}

	if uint32(n) < len {
		return errors.New("unable to write string to buffer")
	}

	// Packet strings are encoded as length then value, where value is padded in
	// multiples of 4 bytes. Make sure we write the padded bytes as well.
	pad := 4 - (len % 4)

	if pad > 0 {
		padBytes := make([]byte, pad)

		n, err = e.buffer.Write(padBytes)
		if err != nil {
			return err
		}

		if uint32(n) < pad {
			return errors.New("unable to write string padding to buffer")
		}
	}

	return nil
}

func (e *Encoder) WriteBytes(data []byte) error {
	n, err := e.buffer.Write(data)
	if err != nil {
		return err
	}

	if n < len(data) {
		return errors.New("unable to write bytes to buffer")
	}

	return nil
}

func (e *Encoder) WriteDateTime(t time.Time) error {
	err := e.WriteUint16(uint16(t.Year()))
	if err != nil {
		return err
	}

	err = e.WriteUint16(uint16(t.Month()))
	if err != nil {
		return err
	}

	// Day of the week, can be 0.
	// Could use t.Weekday() if need-be.
	err = e.WriteUint16(uint16(0))
	if err != nil {
		return err
	}

	err = e.WriteUint16(uint16(t.Day()))
	if err != nil {
		return err
	}

	err = e.WriteUint16(uint16(t.Hour()))
	if err != nil {
		return err
	}

	err = e.WriteUint16(uint16(t.Minute()))
	if err != nil {
		return err
	}

	err = e.WriteUint16(uint16(t.Second()))
	if err != nil {
		return err
	}

	// last field is millisecond
	return e.WriteUint16(0)
}
