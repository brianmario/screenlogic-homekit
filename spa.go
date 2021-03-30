package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type SpaAccessory struct {
	*accessory.Accessory

	heater     *WaterHeaterService
	airBubbles *service.FanV2 // this may need to just be an on/off switch

	client *Client
}

func NewSpaAccessory(client *Client) *SpaAccessory {
	info := accessory.Info{
		Name: "Hot Tub",
		// Model: "",
		Manufacturer: "PentAir",
		// SerialNumber: "",
		// FirmwareRevision: "",
	}

	spa := &SpaAccessory{
		Accessory: accessory.New(info, accessory.TypeHeater),

		client: client,
	}

	spa.heater = NewWaterHeaterService()
	spa.AddService(spa.heater.Service)

	spa.heater.displayUnits.OnValueRemoteGet(spa.client.GetTemperatureDisplayUnits)

	controllerConfig, err := client.getControllerConfig()
	if err != nil {
		return nil
	}

	allowedSpaRange := controllerConfig.AllowedSpaSetPointRange

	min := client.convertTempToHomeKit(uint32(allowedSpaRange.Min))
	max := client.convertTempToHomeKit(uint32(allowedSpaRange.Max))
	step := float64(1.0)

	// The current value must not be less than the minimum we configure, otherwise HomeKit will refuse
	// to pair this accessory.
	spa.heater.heatingThresholdTemperature.SetValue(min)
	spa.heater.heatingThresholdTemperature.SetMinValue(min)
	spa.heater.heatingThresholdTemperature.SetMaxValue(max)
	spa.heater.heatingThresholdTemperature.SetStepValue(step)
	spa.heater.heatingThresholdTemperature.OnValueRemoteGet(spa.client.GetSpaHeatingThresholdTemp)

	spa.heater.HeaterCooler.Active.OnValueRemoteGet(spa.client.GetSpaHeaterActive)

	spa.heater.HeaterCooler.CurrentHeaterCoolerState.OnValueRemoteGet(spa.client.GetSpaCurrentHeatingState)

	spa.heater.HeaterCooler.TargetHeaterCoolerState.OnValueRemoteGet(spa.client.GetSpaTargetHeatingState)

	currentTemp := client.GetCurrentSpaTemp()

	spa.heater.HeaterCooler.CurrentTemperature.SetValue(currentTemp)
	spa.heater.HeaterCooler.CurrentTemperature.OnValueRemoteGet(spa.client.GetCurrentSpaTemp)

	spa.airBubbles = service.NewFanV2()
	spa.AddService(spa.airBubbles.Service)

	return spa
}
