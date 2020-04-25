package main

import (
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type WaterHeaterService struct {
	*service.HeaterCooler

	heatingThresholdTemperature *characteristic.HeatingThresholdTemperature
	displayUnits                *characteristic.TemperatureDisplayUnits
}

func NewWaterHeaterService() *WaterHeaterService {
	svc := &WaterHeaterService{}

	svc.HeaterCooler = service.NewHeaterCooler()

	svc.heatingThresholdTemperature = characteristic.NewHeatingThresholdTemperature()
	svc.AddCharacteristic(svc.heatingThresholdTemperature.Characteristic)

	svc.displayUnits = characteristic.NewTemperatureDisplayUnits()
	svc.AddCharacteristic(svc.displayUnits.Characteristic)

	return svc
}
