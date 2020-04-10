package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
)

type PacketReader struct {
	r io.Reader
}

func NewPacketReader(r io.Reader) *PacketReader {
	return &PacketReader{
		r: r,
	}
}

type PacketHeader struct {
	Sequence uint16
	TypeID   uint16
	Len      uint32
}

type Packet struct {
	PacketHeader

	Data *bytes.Buffer
}

func (pp *PacketReader) ReadPacket(p ReadablePacket) error {
	header := new(PacketHeader)

	err := binary.Read(pp.r, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	// If this packet has a body, read it then pass it along for decoding.
	if header.Len > 0 {
		limitReader := io.LimitReader(pp.r, int64(header.Len))

		dataBuf := new(bytes.Buffer)

		n, err := dataBuf.ReadFrom(limitReader)
		if err != nil {
			return err
		}

		if n != int64(header.Len) {
			return TruncatedPacketError
		}

		return p.Decode(header, dataBuf)
	}

	return p.Decode(header, nil)
}
