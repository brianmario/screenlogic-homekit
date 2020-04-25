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

func (pp *PacketReader) ReadPacket(p ReadablePacket) error {
readAgain:
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

		expectedTypeCode := p.TypeCode()

		if header.TypeID != expectedTypeCode {
			// TODO: deal with whatever this packet type we were sent??
			//
			// What I noticed is that there are some packets the gateway will send us even if
			// we never asked for them. Out of order of the regular request/response cycle.
			// One such packet is the WeatherForcastChanged packet.
			//
			// So for now, since the caller was most likely expecting a different packet here
			// it's probably next in line off the socket buffer, so let's read that now, shall we?
			goto readAgain
		}

		return p.Decode(header, dataBuf)
	}

	return p.Decode(header, nil)
}
