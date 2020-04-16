package screenlogic

import "github.com/brianmario/screenlogic-homekit/screenlogic/protocol"

type PoolStatus struct {
	protocol.PoolStatusResponsePacket
}

func (ps *PoolStatus) AmbientAirTemp() uint32 {
	return ps.AirTemp
}

func (ps *PoolStatus) PoolWater() *protocol.BodyOfWater {
	if len(ps.Bodies) == 0 {
		return nil
	}

	return &ps.Bodies[BodyOfWaterPool]
}

func (ps *PoolStatus) SpaWater() *protocol.BodyOfWater {
	if len(ps.Bodies) < 2 {
		return nil
	}

	return &ps.Bodies[BodyOfWaterSpa]
}

func (ps *PoolStatus) IsReady() bool {
	return ps.OK == 1
}

func (ps *PoolStatus) IsSync() bool {
	return ps.OK == 2
}

func (ps *PoolStatus) IsInServiceMode() bool {
	return ps.OK == 3
}
