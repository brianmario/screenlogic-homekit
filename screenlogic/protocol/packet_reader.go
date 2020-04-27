package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
)

type OOBPacketFn func(header *PacketHeader, data *bytes.Buffer) error

type PacketReader struct {
	r        io.Reader
	callback OOBPacketFn
}

func NewPacketReader(r io.Reader, oobFn OOBPacketFn) *PacketReader {
	return &PacketReader{
		r:        r,
		callback: oobFn,
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
			// What I noticed is that there are some packets the gateway will send us even if
			// we never asked for them. Out of order of the regular request/response cycle.
			// One such packet is the WeatherForcastChanged packet.
			// We'll hand these packets over to the caller so they can decide what to do with them.
			// If this callback returns an error, we stop reading packets here and return that error.
			if pp.callback != nil {
				err = pp.callback(header, dataBuf)
				if err != nil {
					return err
				}
			}

			// As for the packet the caller was most likely expecting here, it's probably next in
			// line off the socket buffer. So let's read that now, shall we?
			goto readAgain
		}

		return p.Decode(header, dataBuf)
	}

	return p.Decode(header, nil)
}
