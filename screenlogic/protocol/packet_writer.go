package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type PacketWriter struct {
	w         io.Writer
	tmpHeader PacketHeader
	packetBuf bytes.Buffer
}

func NewPacketWriter(w io.Writer, startingSequence uint16) *PacketWriter {
	pw := &PacketWriter{
		w: w,
	}

	pw.tmpHeader.Sequence = startingSequence

	return pw
}

func (pw *PacketWriter) WritePacket(p WriteablePacket) error {
	// Always reset our buffer
	defer pw.packetBuf.Reset()

	code, dataBuf, err := p.Encode()
	if err != nil {
		return err
	}

	pw.tmpHeader.TypeID = code

	// Set this packet's length to the *actual* amount of data in its buffer.
	//
	// Also um, this is just WAITING for an unexpected uint64 truncation...
	// ...but we shouldn't see packets larger than UINT32_MAX in this protocol anyway?
	pw.tmpHeader.Len = 0
	if dataBuf != nil {
		pw.tmpHeader.Len = uint32(dataBuf.Len())
	}

	// Next we'll copy to a single contiguous buffer so we can write to the network all
	// at once. The gateway hardware/firmware seems to be sensitive to this, for whatever reason.
	// First the header...
	err = binary.Write(&pw.packetBuf, binary.LittleEndian, pw.tmpHeader)
	if err != nil {
		return err
	}

	// Increment frame sequence
	pw.tmpHeader.Sequence++

	// Now for the actual packet data.
	if dataBuf != nil {
		_, err = dataBuf.WriteTo(&pw.packetBuf)
		if err != nil {
			return err
		}
	}

	expectedWriteLen := pw.packetBuf.Len()

	// Now write the whole thing to the network.
	n, err := pw.packetBuf.WriteTo(pw.w)
	if err != nil {
		return err
	}

	// This probably should never happen, but...
	if n != int64(expectedWriteLen) {
		return errors.New("failed to write packet to the network")
	}

	return nil
}
