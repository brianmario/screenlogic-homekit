package screenlogic

import "github.com/brianmario/screenlogic-homekit/screenlogic/protocol"

type PoolStatus struct {
	protocol.PoolStatusResponsePacket
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
