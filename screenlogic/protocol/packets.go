package protocol

import (
	"bytes"
	"errors"
	"log"
	"net"
)

const (
	LoginFailedCode                           uint16 = 13
	ChallengePacketCode                              = 14
	ChallengePacketResponseCode                      = ChallengePacketCode + 1
	LoginPacketCode                                  = 27
	LoginResponsePacketCode                          = LoginPacketCode + 1
	BadParameterCode                                 = 31
	VersionPacketCode                                = 8120
	VersionResponsePacketCode                        = VersionPacketCode + 1
	ControllerConfigurationPacketCode                = 12532
	ControllerConfigurationResponsePacketCode        = ControllerConfigurationPacketCode + 1
	PoolStatusPacketCode                             = 12526
	PoolStatusResponsePacketCode                     = PoolStatusPacketCode + 1
	SetHeatPointPacketCode                           = 12528
	SetHeatPointResponsePacketCode                   = SetHeatPointPacketCode + 1
	SetHeatModePacketCode                            = 12538
	SetHeatModeResponsePacketCode                    = SetHeatModePacketCode + 1
)

var (
	MalformedPacketErr = errors.New("malformed packet")
	LoginFailedErr     = errors.New("login failed")
)

// WriteablePacket - an interface to give the PacketWriter control over how to write
// arbitrary packet packets to an io.Writer.
type WriteablePacket interface {
	// Encode - should encode this packet's frame data only.
	Encode() (PacketCode uint16, PacketData *bytes.Buffer, err error)
}

// ReadablePacket - an interface to give the PacketReader control over how arbitrary packet
// packets are read from an io.Reader.
type ReadablePacket interface {
	Decode(Header *PacketHeader, PacketData *bytes.Buffer) error
}

// DiscoveryRequestPacketBytes - Technically this data a packet header with the Seq field
// being 1. Presumably, that's because this is always the first packet of the protocol sequence.
//
// This isn't a struct because this packet's data is small, and a constant.
var DiscoveryRequestPacketBytes = []byte("\x01\x00\x00\x00\x00\x00\x00\x00")

// DiscoveryResponsePacket - this is the response we get from the gateway after a
// broadcast request has been sent.
//
// For whatever reason, it doesn't follow the packet framing rules the rest of the
// protocol does. So there's no header.
type DiscoveryResponsePacket struct {
	Type          uint32
	IPAddr        net.IP
	Port          uint16
	GatewayType   uint8
	GatewaySubnet uint8
	GatewayName   string
}

func (drm *DiscoveryResponsePacket) Decode(buf *bytes.Buffer) error {
	decoder := NewDecoder(buf)

	var err error

	drm.Type, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	var ipBytes [4]byte

	n, err := decoder.Read(ipBytes[:])
	if err != nil {
		return err
	}

	if n != 4 {
		return TruncatedPacketError
	}

	drm.IPAddr = net.IPv4(ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])

	drm.Port, err = decoder.ReadUint16()
	if err != nil {
		return err
	}

	drm.GatewayType, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	drm.GatewaySubnet, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	strBuf, err := decoder.ReadTail()
	if err != nil {
		return err
	}

	nullIdx := bytes.Index(strBuf[:], []byte("\x00"))
	if nullIdx == -1 {
		drm.GatewayName = string(strBuf)
	} else {
		drm.GatewayName = string(strBuf[:nullIdx])
	}

	return nil
}

type ChallengePacket struct{}

func (cm *ChallengePacket) Encode() (uint16, *bytes.Buffer, error) {
	return ChallengePacketCode, nil, nil
}

type ChallengePacketResponse struct {
	MacAddr string
}

func (chr *ChallengePacketResponse) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID != ChallengePacketResponseCode {
		return MalformedPacketErr
	}

	decoder := NewDecoder(buf)

	ptr, err := decoder.ReadString()
	if err != nil {
		return nil
	}

	chr.MacAddr = ptr

	return nil
}

type LoginPacket struct {
	Schema         uint32
	ConnectionType uint32
	ClientName     string
	Password       string
	PID            uint32
}

func (lm *LoginPacket) Encode() (uint16, *bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	// Schema
	err := encoder.WriteUint32(lm.Schema)
	if err != nil {
		return 0, nil, err
	}

	// Connection type
	err = encoder.WriteUint32(lm.ConnectionType)
	if err != nil {
		return 0, nil, err
	}

	// Client name
	err = encoder.WriteString(lm.ClientName)
	if err != nil {
		return 0, nil, err
	}

	if len(lm.Password) == 0 {
		// Local connections don't require a password.
		// But we still need to send an empty password field.
		var emptyPass [16]byte

		err = encoder.WriteString(string(emptyPass[:]))
		if err != nil {
			return 0, nil, err
		}
	} else {
		// TODO: encrypt password
		log.Fatal("password encryption not implemented")
	}

	// PID
	// TODO: use our actual pid?
	err = encoder.WriteUint32(lm.PID)
	if err != nil {
		return 0, nil, err
	}

	return LoginPacketCode, buf, nil
}

type LoginResponsePacket struct{}

func (lrm *LoginResponsePacket) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID == LoginFailedCode {
		return LoginFailedErr
	}

	if header.TypeID != LoginResponsePacketCode {
		return MalformedPacketErr
	}

	// TODO: technically there are 16 bytes of something in this packet body
	// we should figure out what they are and deal with them

	return nil
}

type VersionPacket struct{}

func (vm *VersionPacket) Encode() (uint16, *bytes.Buffer, error) {
	return VersionPacketCode, nil, nil
}

type VersionResponsePacket struct {
	Version string
}

func (vm *VersionResponsePacket) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID != VersionResponsePacketCode {
		return MalformedPacketErr
	}

	var err error

	decoder := NewDecoder(buf)

	vm.Version, err = decoder.ReadString()
	if err != nil {
		return nil
	}

	// unknown field 1
	_, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	// unknown field 2
	_, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	// unknown field 3
	_, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	// unknown field 4
	_, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	// unknown field 5
	_, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	// unknown field 6
	_, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	return nil
}

type ControllerConfigurationPacket struct {
	UnknownField1 uint32
	UnknownField2 uint32
}

func (cm *ControllerConfigurationPacket) Encode() (uint16, *bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	// TODO: no idea what these are or what they're for
	err := encoder.WriteUint32(cm.UnknownField1)
	if err != nil {
		return 0, nil, err
	}

	err = encoder.WriteUint32(cm.UnknownField2)
	if err != nil {
		return 0, nil, err
	}

	return ControllerConfigurationPacketCode, buf, nil
}

type SetPoint struct {
	Min uint8
	Max uint8
}

// TODO: In other libraries this number is hard coded to 8. Why?
const maxPumpCount uint8 = 8

type ControllerConfigurationResponsePacket struct {
	ControllerID       uint32
	PoolSetPoint       SetPoint
	SpaSetPoint        SetPoint
	IsCelcius          bool
	ControllerType     uint8
	HardwareType       uint8
	ControllerBuffer   uint8
	EquipmentFlags     uint32
	DefaultCircuitName string
	Circuits           []ControllerCircuit
	Colors             []Color
	Pumps              [maxPumpCount]Pump
	InterfaceTabFlags  uint32
	ShowAlarms         bool
}

type ControllerCircuit struct {
	ID            uint32
	Name          string
	NameIndex     uint8
	Function      uint8
	Interface     uint8
	Flags         uint8
	ColorSet      uint8
	ColorPosition uint8
	ColorStagger  uint8
	DeviceID      uint8
	DefaultRT     uint16
}

type Color struct {
	Name  string
	Red   uint32
	Green uint32
	Blue  uint32
}

type Pump struct {
	Data uint8
}

func (vm *ControllerConfigurationResponsePacket) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID != ControllerConfigurationResponsePacketCode {
		return MalformedPacketErr
	}

	var err error

	decoder := NewDecoder(buf)

	vm.ControllerID, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	vm.PoolSetPoint.Min, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.PoolSetPoint.Max, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.SpaSetPoint.Min, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.SpaSetPoint.Max, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.IsCelcius, err = decoder.ReadBool()
	if err != nil {
		return err
	}

	vm.ControllerType, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.HardwareType, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.ControllerBuffer, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.EquipmentFlags, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	vm.DefaultCircuitName, err = decoder.ReadString()
	if err != nil {
		return err
	}

	circuitCount, err := decoder.ReadUint32()
	if err != nil {
		return err
	}

	if circuitCount > 0 {
		vm.Circuits = make([]ControllerCircuit, circuitCount)

		var i uint32 = 0
		for ; i < circuitCount; i++ {
			circuit := &vm.Circuits[i]

			circuit.ID, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			circuit.Name, err = decoder.ReadString()
			if err != nil {
				return err
			}

			circuit.NameIndex, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.Function, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.Interface, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.Flags, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.ColorSet, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.ColorPosition, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.ColorStagger, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.DeviceID, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.DefaultRT, err = decoder.ReadUint16()
			if err != nil {
				return err
			}

			// scoot our read pointer 2 spots
			// TODO: what field is this?
			_, err = decoder.ReadUint16()
			if err != nil {
				return err
			}
		}
	}

	colorCount, err := decoder.ReadUint32()
	if err != nil {
		return err
	}

	if colorCount > 0 {
		vm.Colors = make([]Color, colorCount)

		var i uint32 = 0
		for ; i < colorCount; i++ {
			color := &vm.Colors[i]

			color.Name, err = decoder.ReadString()
			if err != nil {
				return err
			}

			color.Red, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			color.Green, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			color.Blue, err = decoder.ReadUint32()
			if err != nil {
				return err
			}
		}
	}

	var i uint8 = 0
	for ; i < maxPumpCount; i++ {
		pump := &vm.Pumps[i]

		pump.Data, err = decoder.ReadUint8()
		if err != nil {
			return err
		}
	}

	vm.InterfaceTabFlags, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	vm.ShowAlarms, err = decoder.ReadBool()
	if err != nil {
		return err
	}

	return nil
}

type PoolStatusPacket struct {
	UnknownField uint32
}

func (psp *PoolStatusPacket) Encode() (uint16, *bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	// TODO: no idea what this is or what it's for
	err := encoder.WriteUint32(psp.UnknownField)
	if err != nil {
		return 0, nil, err
	}

	return PoolStatusPacketCode, buf, nil
}

type PoolStatusResponsePacket struct {
	OK           uint32
	FreezeMode   uint8
	Remotes      uint8
	PoolDelay    uint8
	SpaDelay     uint8
	CleanerDelay uint8
	WhoKnows     [3]byte
	AirTemp      uint32
	Bodies       []Body
	Circuits     []PoolCircuit
	Chemistry    struct {
		PH           float32
		ORP          float32
		Saturation   float32
		SaltPPM      uint32
		PHTankLevel  uint32
		ORPTankLevel uint32
		Alarms       uint32
	}
}

type Body struct {
	Type         uint32 // other libraries appear to force this to be with in 0 or 1
	CurrentTemp  uint32
	HeaterStatus uint32
	HeatSetPoint uint32
	CoolSetPoint uint32
	HeatMode     uint32
}

type PoolCircuit struct {
	ID            uint32
	ValveState    uint32
	ColorSet      uint8
	ColorPosition uint8
	ColorStagger  uint8
	Delay         uint8
}

func (psrp *PoolStatusResponsePacket) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID != PoolStatusResponsePacketCode {
		return MalformedPacketErr
	}

	var err error

	decoder := NewDecoder(buf)

	psrp.OK, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	psrp.FreezeMode, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	psrp.Remotes, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	psrp.PoolDelay, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	psrp.SpaDelay, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	psrp.CleanerDelay, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	// ignore these 3 bytes?
	err = decoder.CopyBytes(psrp.WhoKnows[:])
	if err != nil {
		return err
	}

	psrp.AirTemp, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	// bodies of water??
	// other libraries seem to force this to be at most 2
	bodyCount, err := decoder.ReadUint32()
	if err != nil {
		return err
	}

	// TODO: another library does this, but I'm not sure why?
	// We'll skip it and see what happens ;)
	// if bodyCount > 2 {
	// 	bodyCount = 2
	// }

	if bodyCount > 0 {
		psrp.Bodies = make([]Body, bodyCount)

		var i uint32 = 0
		for ; i < bodyCount; i++ {
			body := &psrp.Bodies[i]

			body.Type, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			body.CurrentTemp, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			body.HeaterStatus, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			body.HeatSetPoint, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			body.CoolSetPoint, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			body.HeatMode, err = decoder.ReadUint32()
			if err != nil {
				return err
			}
		}
	} else {
		// TODO: this should probably error?
	}

	circuitCount, err := decoder.ReadUint32()
	if err != nil {
		return err
	}

	if circuitCount > 0 {
		psrp.Circuits = make([]PoolCircuit, circuitCount)

		var i uint32 = 0
		for ; i < circuitCount; i++ {
			circuit := &psrp.Circuits[i]

			circuit.ID, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			circuit.ValveState, err = decoder.ReadUint32()
			if err != nil {
				return err
			}

			circuit.ColorSet, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.ColorPosition, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.ColorStagger, err = decoder.ReadUint8()
			if err != nil {
				return err
			}

			circuit.Delay, err = decoder.ReadUint8()
			if err != nil {
				return err
			}
		}
	}

	ph, err := decoder.ReadUint32()
	if err != nil {
		return err
	}

	psrp.Chemistry.PH = float32(ph) / 100

	orp, err := decoder.ReadUint32()
	if err != nil {
		return err
	}

	psrp.Chemistry.ORP = float32(orp) / 100

	saturation, err := decoder.ReadUint32()
	if err != nil {
		return err
	}

	psrp.Chemistry.Saturation = float32(saturation) / 100

	psrp.Chemistry.SaltPPM, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	psrp.Chemistry.PHTankLevel, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	psrp.Chemistry.ORPTankLevel, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	psrp.Chemistry.Alarms, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	return nil
}

type SetHeatPointPacket struct {
	ControllerIdx uint32
	BodyType      uint32
	Temperature   uint32
}

func (shp *SetHeatPointPacket) Encode() (uint16, *bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	err := encoder.WriteUint32(shp.ControllerIdx)
	if err != nil {
		return 0, nil, err
	}

	err = encoder.WriteUint32(shp.BodyType)
	if err != nil {
		return 0, nil, err
	}

	err = encoder.WriteUint32(shp.Temperature)
	if err != nil {
		return 0, nil, err
	}

	return SetHeatPointPacketCode, buf, nil
}

type SetHeatPointResponsePacket struct{}

func (shpr *SetHeatPointResponsePacket) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID != SetHeatPointResponsePacketCode {
		return MalformedPacketErr
	}

	// this presumably has no fields?

	return nil
}

type SetHeatModePacket struct {
	ControllerIdx uint32
	BodyType      uint32
	Mode          uint32
}

func (shm *SetHeatModePacket) Encode() (uint16, *bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	err := encoder.WriteUint32(shm.ControllerIdx)
	if err != nil {
		return 0, nil, err
	}

	err = encoder.WriteUint32(shm.BodyType)
	if err != nil {
		return 0, nil, err
	}

	err = encoder.WriteUint32(shm.Mode)
	if err != nil {
		return 0, nil, err
	}

	return SetHeatModePacketCode, buf, nil
}

type SetHeatModeResponsePacket struct{}

func (shmr *SetHeatModeResponsePacket) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID != SetHeatModeResponsePacketCode {
		return MalformedPacketErr
	}

	// this presumably has no fields?

	return nil
}
