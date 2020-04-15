package main

import (
	"log"

	"github.com/brianmario/screenlogic-homekit/screenlogic"
)

func main() {
	gateway, err := screenlogic.DiscoverGateway()
	if err != nil {
		log.Fatal(err)
	}

	err = gateway.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer gateway.Close()

	clientName := "screenlogic-homekit"

	err = gateway.Login(clientName)
	if err != nil {
		log.Fatal(err)
	}

	_, err = gateway.Version()
	if err != nil {
		log.Fatal(err)
	}

	_, err = gateway.ControllerConfig()
	if err != nil {
		log.Fatal(err)
	}

	_, err = gateway.PoolStatus()
	if err != nil {
		log.Fatal(err)
	}

	// I only have one pool controller.
	// I'm not sure where a list of them is returned, or where you'd get the
	// index of a controller, but I can assume 0 here.
	var controllerIdx uint32 = 0

	// This will eventually come from HomeKit
	var setTemp uint32 = 80

	poolID := screenlogic.Pool

	err = gateway.SetTemperature(controllerIdx, poolID, setTemp)
	if err != nil {
		log.Fatal(err)
	}

	heatMode := screenlogic.HeatModeOff

	err = gateway.SetHeatMode(controllerIdx, poolID, heatMode)
	if err != nil {
		log.Fatal(err)
	}
}
