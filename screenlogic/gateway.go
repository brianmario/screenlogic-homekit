package screenlogic

import (
	"bytes"
	"fmt"
	"net"

	"github.com/brianmario/screenlogic-homekit/screenlogic/protocol"
)

type Gateway struct {
	client         net.Conn
	packetSequence uint16
	packetReader   *protocol.PacketReader
	packetWriter   *protocol.PacketWriter

	IP      net.IP
	Port    uint16
	Type    uint8
	Subnet  uint8
	Name    string
	MacAddr string
}

const DiscoveryPort = 1444

func DiscoverGateway() (*Gateway, error) {
	addr := fmt.Sprintf("255.255.255.255:%d", DiscoveryPort)

	broadcastAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}

	listenSock, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, err
	}
	defer listenSock.Close()

	_, err = listenSock.WriteTo(protocol.DiscoveryRequestPacketBytes, broadcastAddr)
	if err != nil {
		return nil, err
	}

	// This next packet doesn't follow the packet framing spec the rest of the
	// protocol does, so we have to decode this special snowflake without the PacketReader
	var resp protocol.DiscoveryResponsePacket

	var tmpPacketBuf [40]byte

	n, err := listenSock.Read(tmpPacketBuf[:])
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(tmpPacketBuf[:n])

	err = resp.Decode(buffer)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		IP:     resp.IPAddr,
		Port:   resp.Port,
		Type:   resp.GatewayType,
		Subnet: resp.GatewaySubnet,
		Name:   resp.GatewayName,
	}, nil
}

func (g *Gateway) Connect() error {
	var err error

	g.client, err = net.Dial("tcp4", fmt.Sprintf("%s:%d", g.IP, g.Port))
	if err != nil {
		return err
	}

	// Setup packet processing
	g.packetReader = protocol.NewPacketReader(g.client)
	g.packetWriter = protocol.NewPacketWriter(g.client, 2)

	// Again, this packet doesn't follow the packet framing spec, so we'll just write
	// directly to the socket.
	_, err = g.client.Write([]byte("CONNECTSERVERHOST\r\n\r\n"))
	if err != nil {
		return err
	}

	var challenge protocol.ChallengePacket

	err = g.packetWriter.WritePacket(&challenge)
	if err != nil {
		return err
	}

	var resp protocol.ChallengePacketResponse

	err = g.packetReader.ReadPacket(&resp)
	if err != nil {
		return err
	}

	g.MacAddr = resp.MacAddr

	return nil
}

func (g *Gateway) Login(clientName string) error {
	var req protocol.LoginPacket

	req.Schema = 348       // this was picked up from another OSS client
	req.ConnectionType = 0 // so was this
	req.ClientName = clientName
	req.Password = "" // TODO: allow passing password
	req.PID = 2       // TODO: use our actual PID?

	err := g.packetWriter.WritePacket(&req)
	if err != nil {
		return err
	}

	var resp protocol.LoginResponsePacket

	err = g.packetReader.ReadPacket(&resp)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gateway) Version() (string, error) {
	var req protocol.VersionPacket

	err := g.packetWriter.WritePacket(&req)
	if err != nil {
		return "", err
	}

	var resp protocol.VersionResponsePacket

	err = g.packetReader.ReadPacket(&resp)
	if err != nil {
		return "", err
	}

	return resp.Version, nil
}

func (g *Gateway) ControllerConfig() (*ControllerConfiguration, error) {
	var req protocol.ControllerConfigurationPacket

	err := g.packetWriter.WritePacket(&req)
	if err != nil {
		return nil, err
	}

	resp := &ControllerConfiguration{}

	err = g.packetReader.ReadPacket(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *Gateway) PoolStatus() (*PoolStatus, error) {
	var req protocol.PoolStatusPacket

	err := g.packetWriter.WritePacket(&req)
	if err != nil {
		return nil, err
	}

	resp := &PoolStatus{}

	err = g.packetReader.ReadPacket(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type BodyOfWater uint32

const (
	Pool BodyOfWater = iota
	Spa
)

func (g *Gateway) SetTemperature(controllerIdx uint32, bodyType BodyOfWater, temperature uint32) error {
	var req protocol.SetHeatPointPacket

	req.ControllerIdx = controllerIdx
	req.BodyType = uint32(bodyType)
	req.Temperature = temperature

	err := g.packetWriter.WritePacket(&req)
	if err != nil {
		return err
	}

	var resp protocol.SetHeatPointResponsePacket

	err = g.packetReader.ReadPacket(&resp)
	if err != nil {
		return err
	}

	return nil
}

type HeatMode uint32

const (
	HeatModeOff HeatMode = iota
	HeatModeSolarOnly
	HeatModeSolarPreferred
	HeatModeOn
	HeatModeUnchanged
)

func (g *Gateway) SetHeatMode(controllerIdx uint32, bodyType BodyOfWater, mode HeatMode) error {
	var req protocol.SetHeatModePacket

	req.ControllerIdx = controllerIdx
	req.BodyType = uint32(bodyType)
	req.Mode = uint32(mode)

	err := g.packetWriter.WritePacket(&req)
	if err != nil {
		return err
	}

	var resp protocol.SetHeatModeResponsePacket

	err = g.packetReader.ReadPacket(&resp)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gateway) Close() {
	g.client.Close()
}
