package main

import (
	"flag"
	"strings"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"
)

var pinCode string

const pinCodeDefault = "00102003"

func main() {
	// Enable debug logging in github.com/brutella/hc as well as in this codebase.
	//
	// log.Debug.Enable()

	flag.StringVar(&pinCode, "pin", pinCodeDefault, "homekit pin code to use for this accessory")

	flag.Parse()

	client, err := NewConnectedClient("screenlogic-homekit")
	if err != nil {
		log.Debug.Fatal(err)
	}

	airTempInfo := accessory.Info{
		Manufacturer: "Pentair",
		Model:        "ScreenLogic",
		Name:         "Ambient Air Temperature",
	}

	currentAirTemp, err := client.GetAirTemperature()
	if err != nil {
		log.Debug.Fatal(err)
	}

	airTemp := accessory.NewTemperatureSensor(airTempInfo, currentAirTemp, -40, 150, 0.25)

	pool := NewPoolAccessory(client)

	spa := NewSpaAccessory(client)

	pwConfig := hc.Config{Pin: pinCode}

	gatewayName := client.GetGatewayName()

	// Make the name safe for HomeKit by removing the space and the ':' char.
	// Also call it ScreenLogic instead of the more generic Pentair name.
	gatewayName = strings.Replace(gatewayName, "Pentair: ", "ScreenLogic-", 1)

	gatewayVersion, err := client.GetGatewayVersion()
	if err != nil {
		log.Debug.Fatal(err)
	}

	bridgeInfo := accessory.Info{
		Manufacturer:     "Pentair",
		Model:            "ScreenLogic",
		Name:             gatewayName,
		FirmwareRevision: gatewayVersion,
	}

	bridge := accessory.NewBridge(bridgeInfo)

	// NOTE: the first accessory in the list acts as the bridge, while the rest will be linked to it
	t, err := hc.NewIPTransport(pwConfig, bridge.Accessory, airTemp.Accessory, pool.Accessory, spa.Accessory)
	if err != nil {
		log.Debug.Panic(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	t.Start()
}
