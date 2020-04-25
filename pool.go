package main

import (
	"github.com/brutella/hc/accessory"
)

type PoolAccessory struct {
	*accessory.Accessory

	heater *WaterHeaterService

	client *Client
}

func NewPoolAccessory(client *Client) *PoolAccessory {
	info := accessory.Info{
		Name: "Pool",
		// Model: "",
		Manufacturer: "PentAir",
		// SerialNumber: "",
		// FirmwareRevision: "",
	}

	pool := &PoolAccessory{
		Accessory: accessory.New(info, accessory.TypeHeater),

		client: client,
	}

	pool.heater = NewWaterHeaterService()
	pool.AddService(pool.heater.Service)

	pool.heater.displayUnits.OnValueRemoteGet(pool.client.GetTemperatureDisplayUnits)

	controllerConfig, err := client.getControllerConfig()
	if err != nil {
		return nil
	}

	allowedPoolRange := controllerConfig.AllowedPoolSetPointRange

	min := client.convertTempToHomeKit(uint32(allowedPoolRange.Min))
	max := client.convertTempToHomeKit(uint32(allowedPoolRange.Max))
	step := float64(1.0)

	// The current value must not be less than the minimum we configure, otherwise HomeKit will refuse
	// to pair this accessory.
	pool.heater.heatingThresholdTemperature.SetValue(min)
	pool.heater.heatingThresholdTemperature.SetMinValue(min)
	pool.heater.heatingThresholdTemperature.SetMaxValue(max)
	pool.heater.heatingThresholdTemperature.SetStepValue(step)
	pool.heater.heatingThresholdTemperature.OnValueRemoteGet(pool.client.GetPoolHeatingThresholdTemp)

	pool.heater.HeaterCooler.Active.OnValueRemoteGet(pool.client.GetPoolHeaterActive)

	pool.heater.HeaterCooler.CurrentHeaterCoolerState.OnValueRemoteGet(pool.client.GetPoolCurrentHeatingState)

	pool.heater.HeaterCooler.TargetHeaterCoolerState.OnValueRemoteGet(pool.client.GetPoolTargetHeatingState)

	currentTemp := client.GetCurrentPoolTemp()

	pool.heater.HeaterCooler.CurrentTemperature.SetValue(currentTemp)
	pool.heater.HeaterCooler.CurrentTemperature.OnValueRemoteGet(pool.client.GetCurrentPoolTemp)

	return pool
}
