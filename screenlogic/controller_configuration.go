package screenlogic

import "github.com/brianmario/screenlogic-homekit/screenlogic/protocol"

type ControllerConfiguration struct {
	protocol.ControllerConfigurationResponsePacket
}

func (cc *ControllerConfiguration) HasSolar() bool {
	return (cc.EquipmentFlags & 0x1) != 0
}

func (cc *ControllerConfiguration) HasSolarAsHeatpump() bool {
	return (cc.EquipmentFlags & 0x2) != 0
}

func (cc *ControllerConfiguration) HasChlorinator() bool {
	return (cc.EquipmentFlags & 0x4) != 0
}

func (cc *ControllerConfiguration) HasCooling() bool {
	return (cc.EquipmentFlags & 0x800) != 0
}

func (cc *ControllerConfiguration) HasIntellichem() bool {
	return (cc.EquipmentFlags & 0x8000) != 0
}

func (cc *ControllerConfiguration) IsEasyTouch() bool {
	return cc.ControllerType == 14 || cc.ControllerType == 13
}

func (cc *ControllerConfiguration) IsIntelliTouch() bool {
	return cc.ControllerType != 14 && cc.ControllerType != 13 && cc.ControllerType != 10
}

func (cc *ControllerConfiguration) IsEasyTouchLite() bool {
	return cc.ControllerType == 13 && (cc.HardwareType&0x4) != 0
}

func (cc *ControllerConfiguration) IsDualBody() bool {
	return cc.ControllerType == 5
}

func (cc *ControllerConfiguration) IsChem2() bool {
	return cc.ControllerType == 252 && cc.HardwareType == 2
}
