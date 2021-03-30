package protocol

import (
	"bytes"
	"errors"
	"log"
	"net"
	"time"
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
	WeatherForcastChangedCode                        = 9806
	HistoryDataResponsePacketCode                    = 12502
	ControllerConfigurationPacketCode                = 12532
	ControllerConfigurationResponsePacketCode        = ControllerConfigurationPacketCode + 1
	PoolStatusPacketCode                             = 12526
	PoolStatusResponsePacketCode                     = PoolStatusPacketCode + 1
	SetHeatPointPacketCode                           = 12528
	SetHeatPointResponsePacketCode                   = SetHeatPointPacketCode + 1
	HistoryPacketCode                                = 12534
	HistoryPacketResponseCode                        = HistoryPacketCode + 1
	SetHeatModePacketCode                            = 12538
	SetHeatModeResponsePacketCode                    = SetHeatModePacketCode + 1
)

var (
	MalformedPacketErr = errors.New("malformed packet")
	LoginFailedErr     = errors.New("login failed")
)

type IdentifiablePacket interface {
	TypeCode() uint16
}

// WriteablePacket - an interface to give the PacketWriter control over how to write
// arbitrary packet packets to an io.Writer.
type WriteablePacket interface {
	IdentifiablePacket

	// Encode - should encode this packet's frame data only.
	Encode() (PacketData *bytes.Buffer, err error)
}

// ReadablePacket - an interface to give the PacketReader control over how arbitrary packet
// packets are read from an io.Reader.
type ReadablePacket interface {
	IdentifiablePacket

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

func (cm *ChallengePacket) TypeCode() uint16 {
	return ChallengePacketCode
}

func (cm *ChallengePacket) Encode() (*bytes.Buffer, error) {
	return nil, nil
}

type ChallengePacketResponse struct {
	MacAddr string
}

func (cm *ChallengePacketResponse) TypeCode() uint16 {
	return ChallengePacketResponseCode
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

func (lm *LoginPacket) TypeCode() uint16 {
	return LoginPacketCode
}

func (lm *LoginPacket) Encode() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	// Schema
	err := encoder.WriteUint32(lm.Schema)
	if err != nil {
		return nil, err
	}

	// Connection type
	err = encoder.WriteUint32(lm.ConnectionType)
	if err != nil {
		return nil, err
	}

	// Client name
	err = encoder.WriteString(lm.ClientName)
	if err != nil {
		return nil, err
	}

	if len(lm.Password) == 0 {
		// Local connections don't require a password.
		// But we still need to send an empty password field.
		var emptyPass [16]byte

		err = encoder.WriteString(string(emptyPass[:]))
		if err != nil {
			return nil, err
		}
	} else {
		// TODO: encrypt password
		log.Fatal("password encryption not implemented")
	}

	// PID
	// TODO: use our actual pid?
	err = encoder.WriteUint32(lm.PID)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

type LoginResponsePacket struct{}

func (lm *LoginResponsePacket) TypeCode() uint16 {
	return LoginResponsePacketCode
}

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

func (vp *VersionPacket) TypeCode() uint16 {
	return VersionPacketCode
}

func (vm *VersionPacket) Encode() (*bytes.Buffer, error) {
	return nil, nil
}

type VersionResponsePacket struct {
	Version string
}

func (vrp *VersionResponsePacket) TypeCode() uint16 {
	return VersionResponsePacketCode
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

func (ccp *ControllerConfigurationPacket) TypeCode() uint16 {
	return ControllerConfigurationPacketCode
}

func (cm *ControllerConfigurationPacket) Encode() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	// TODO: no idea what these are or what they're for
	err := encoder.WriteUint32(cm.UnknownField1)
	if err != nil {
		return nil, err
	}

	err = encoder.WriteUint32(cm.UnknownField2)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

type SetPoint struct {
	Min uint8
	Max uint8
}

// TODO: In other libraries this number is hard coded to 8. Why?
const maxPumpCount uint8 = 8

type ControllerConfigurationResponsePacket struct {
	ControllerID             uint32
	AllowedPoolSetPointRange SetPoint
	AllowedSpaSetPointRange  SetPoint
	IsCelcius                bool
	ControllerType           uint8
	HardwareType             uint8
	ControllerBuffer         uint8
	EquipmentFlags           uint32
	DefaultCircuitName       string
	Circuits                 []ControllerCircuit
	Colors                   []Color
	Pumps                    [maxPumpCount]Pump
	InterfaceTabFlags        uint32
	ShowAlarms               bool
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

func (ccrp *ControllerConfigurationResponsePacket) TypeCode() uint16 {
	return ControllerConfigurationResponsePacketCode
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

	vm.AllowedPoolSetPointRange.Min, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.AllowedPoolSetPointRange.Max, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.AllowedSpaSetPointRange.Min, err = decoder.ReadUint8()
	if err != nil {
		return err
	}

	vm.AllowedSpaSetPointRange.Max, err = decoder.ReadUint8()
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

func (psp *PoolStatusPacket) TypeCode() uint16 {
	return PoolStatusPacketCode
}

func (psp *PoolStatusPacket) Encode() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	// TODO: no idea what this is or what it's for
	err := encoder.WriteUint32(psp.UnknownField)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

type PoolStatusResponsePacket struct {
	OK           uint32 // TODO: figure out a better name for this
	FreezeMode   uint8
	Remotes      uint8
	PoolDelay    uint8
	SpaDelay     uint8
	CleanerDelay uint8
	WhoKnows     [3]byte
	AirTemp      uint32
	Bodies       []BodyOfWater
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

type BodyOfWater struct {
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

func (psrp *PoolStatusResponsePacket) TypeCode() uint16 {
	return PoolStatusResponsePacketCode
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
		psrp.Bodies = make([]BodyOfWater, bodyCount)

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

func (shpp *SetHeatPointPacket) TypeCode() uint16 {
	return SetHeatPointPacketCode
}

func (shp *SetHeatPointPacket) Encode() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	err := encoder.WriteUint32(shp.ControllerIdx)
	if err != nil {
		return nil, err
	}

	err = encoder.WriteUint32(shp.BodyType)
	if err != nil {
		return nil, err
	}

	err = encoder.WriteUint32(shp.Temperature)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

type SetHeatPointResponsePacket struct{}

func (shprp *SetHeatPointResponsePacket) TypeCode() uint16 {
	return SetHeatPointResponsePacketCode
}

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

func (shmp *SetHeatModePacket) TypeCode() uint16 {
	return SetHeatModePacketCode
}

func (shm *SetHeatModePacket) Encode() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	err := encoder.WriteUint32(shm.ControllerIdx)
	if err != nil {
		return nil, err
	}

	err = encoder.WriteUint32(shm.BodyType)
	if err != nil {
		return nil, err
	}

	err = encoder.WriteUint32(shm.Mode)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

type SetHeatModeResponsePacket struct{}

func (shmrp *SetHeatModeResponsePacket) TypeCode() uint16 {
	return SetHeatModeResponsePacketCode
}

func (shmr *SetHeatModeResponsePacket) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID != SetHeatModeResponsePacketCode {
		return MalformedPacketErr
	}

	// this presumably has no fields?

	return nil
}

type HistoryPacket struct {
	ControllerIndex uint32 // use 0
	Start           time.Time
	End             time.Time
	SenderID        uint32 // not used, use 0
}

func (hp *HistoryPacket) TypeCode() uint16 {
	return HistoryPacketCode
}

func (hp *HistoryPacket) Encode() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)

	err := encoder.WriteUint32(hp.ControllerIndex)
	if err != nil {
		return nil, err
	}

	err = encoder.WriteDateTime(hp.Start)
	if err != nil {
		return nil, err
	}

	err = encoder.WriteDateTime(hp.End)
	if err != nil {
		return nil, err
	}

	err = encoder.WriteUint32(hp.SenderID)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

type HistoryResponsePacket struct{}

func (hrp *HistoryResponsePacket) TypeCode() uint16 {
	return HistoryPacketResponseCode
}

func (hrp *HistoryResponsePacket) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID != HistoryPacketResponseCode {
		return MalformedPacketErr
	}

	// this presumably has no fields?

	return nil
}

type HistoryEvent struct {
	Timestamp time.Time
	Temp      uint32
}

type HistoryDataResponsePacket struct {
	OutsideTemps   []HistoryEvent
	PoolWaterTemps []HistoryEvent
}

func (hrp *HistoryDataResponsePacket) TypeCode() uint16 {
	return HistoryDataResponsePacketCode
}

func (hrp *HistoryDataResponsePacket) Decode(header *PacketHeader, buf *bytes.Buffer) error {
	if header.TypeID != HistoryDataResponsePacketCode {
		return MalformedPacketErr
	}

	decoder := NewDecoder(buf)

	// outside temp?
	numEvents, err := decoder.ReadUint32()
	if err != nil {
		return err
	}

	hrp.OutsideTemps = make([]HistoryEvent, numEvents)

	for i := uint32(0); i < numEvents; i++ {
		event := &hrp.OutsideTemps[i]

		event.Timestamp, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}

		event.Temp, err = decoder.ReadUint32()
		if err != nil {
			return err
		}
	}
	// outside temp?

	// pool water temp?
	numEvents, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	hrp.PoolWaterTemps = make([]HistoryEvent, numEvents)

	for i := uint32(0); i < numEvents; i++ {
		event := &hrp.PoolWaterTemps[i]

		event.Timestamp, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}

		event.Temp, err = decoder.ReadUint32()
		if err != nil {
			return err
		}
	}
	// pool water temp?

	// no idea, last hot tub temp?
	numEvents, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	for i := uint32(0); i < numEvents; i++ {
		_, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}

		_, err = decoder.ReadUint32()
		if err != nil {
			return err
		}
	}
	// no idea, last hot tub temp?

	// not sure, seems to be 0 length
	numEvents, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	for i := uint32(0); i < numEvents; i++ {
		_, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}

		_, err = decoder.ReadUint32()
		if err != nil {
			return err
		}
	}
	// not sure, seems to be 0 length

	// not sure, seems to be 0 length
	numEvents, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	for i := uint32(0); i < numEvents; i++ {
		_, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}

		_, err = decoder.ReadUint32()
		if err != nil {
			return err
		}
	}
	// not sure, seems to be 0 length

	// timestamps only, maybe some sort of state change record?
	numEvents, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	for i := uint32(0); i < numEvents; i++ {
		_, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}
	}
	// timestamps only, maybe some sort of state change record?

	// there appear to be the same number of timestamps from the previous numEvents field
	for i := uint32(0); i < numEvents; i++ {
		_, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}
	}
	// there appear to be the same number of timestamps from the previous numEvents field

	// not sure, seems to be 0 length
	numEvents, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	for i := uint32(0); i < numEvents; i++ {
		_, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}

		_, err = decoder.ReadUint32()
		if err != nil {
			return err
		}
	}
	// not sure, seems to be 0 length

	// not sure, seems to be 0 length
	numEvents, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	for i := uint32(0); i < numEvents; i++ {
		_, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}

		_, err = decoder.ReadUint32()
		if err != nil {
			return err
		}
	}
	// not sure, seems to be 0 length

	// not sure, seems to be 0 length
	numEvents, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	for i := uint32(0); i < numEvents; i++ {
		_, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}

		_, err = decoder.ReadUint32()
		if err != nil {
			return err
		}
	}
	// not sure, seems to be 0 length

	// not sure, seems to be 0 length
	numEvents, err = decoder.ReadUint32()
	if err != nil {
		return err
	}

	for i := uint32(0); i < numEvents; i++ {
		_, err = decoder.ReadDateTime()
		if err != nil {
			return err
		}

		_, err = decoder.ReadUint32()
		if err != nil {
			return err
		}
	}
	// not sure, seems to be 0 length

	return nil
}
